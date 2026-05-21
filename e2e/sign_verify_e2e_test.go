// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build darwin && e2e

package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/apple/spctl"
	"github.com/convergent-systems-co/macforge/internal/config"
	"github.com/convergent-systems-co/macforge/internal/keychain"
	"github.com/convergent-systems-co/macforge/internal/signing"
	"github.com/convergent-systems-co/macforge/internal/verify"
)

func buildHelloApp(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("clang not available; cannot build HelloApp")
	}

	dir := t.TempDir()
	appDir := filepath.Join(dir, "HelloApp.app", "Contents", "MacOS")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	const infoPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleIdentifier</key>           <string>com.convergent.macforge.hello</string>
  <key>CFBundleName</key>                 <string>HelloApp</string>
  <key>CFBundleExecutable</key>           <string>HelloApp</string>
  <key>CFBundleVersion</key>              <string>1</string>
  <key>CFBundleShortVersionString</key>   <string>1.0</string>
  <key>CFBundlePackageType</key>          <string>APPL</string>
</dict>
</plist>
`
	plistPath := filepath.Join(dir, "HelloApp.app", "Contents", "Info.plist")
	if err := os.WriteFile(plistPath, []byte(infoPlist), 0o644); err != nil {
		t.Fatalf("write Info.plist: %v", err)
	}

	const cSrc = `#include <stdio.h>
int main(void) { puts("hello, macforge"); return 0; }
`
	srcPath := filepath.Join(dir, "hello.c")
	if err := os.WriteFile(srcPath, []byte(cSrc), 0o644); err != nil {
		t.Fatalf("write hello.c: %v", err)
	}

	binPath := filepath.Join(appDir, "HelloApp")
	out, err := exec.Command("clang", "-arch", "arm64", "-o", binPath, srcPath).CombinedOutput()
	if err != nil {
		t.Fatalf("clang failed: %v\n%s", err, out)
	}

	return filepath.Join(dir, "HelloApp.app")
}

func TestE2E_Sign_Verify_HelloApp(t *testing.T) {
	requireEnv(t,
		"MACFORGE_E2E_TEAM",
		"MACFORGE_E2E_KEYCHAIN_PASSWORD",
		"MACFORGE_E2E_CERT_PATH",
	)

	team := os.Getenv("MACFORGE_E2E_TEAM")
	t.Setenv("MACFORGE_TEST_PW", os.Getenv("MACFORGE_E2E_KEYCHAIN_PASSWORD"))

	appPath := buildHelloApp(t)

	r := apple.NewExecRunner(nil)
	sec := security.New(r)
	m := keychain.NewManager(sec)
	name := keychain.DefaultName(team, "e2e-sign")
	ctx := context.Background()

	_ = m.Delete(ctx, name)
	if err := m.Create(ctx, keychain.CreateOptions{
		Name:            name,
		SecretRef:       "env:MACFORGE_TEST_PW",
		LockOnSleep:     false,
		LockTimeoutSecs: 600,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = m.Delete(context.Background(), name) })

	if err := m.Unlock(ctx, name, "env:MACFORGE_TEST_PW"); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if err := sec.Import(ctx, os.Getenv("MACFORGE_E2E_CERT_PATH"), name, os.Getenv("MACFORGE_E2E_P12_PASSWORD")); err != nil {
		t.Fatalf("Import: %v", err)
	}

	cfg := &config.Config{
		Version:  1,
		Team:     team,
		Keychain: config.KeychainConfig{Name: name},
		Sign: config.SignConfig{
			HardenedRuntime: true,
			Timestamp:       true,
		},
	}

	signer := signing.New(codesign.New(r), sec)
	if _, err := signer.Sign(ctx, cfg, appPath); err != nil {
		t.Fatalf("Sign: %v", err)
	}

	verifier := verify.New(codesign.New(r), spctl.New(r))
	res, err := verifier.Verify(ctx, appPath, spctl.AssessTypeExec)
	if err != nil && !res.Codesign.OK {
		t.Fatalf("Verify codesign: %v", err)
	}
	if !res.Codesign.OK {
		t.Fatalf("Codesign.OK = false; stderr: %s", res.Codesign.Stderr)
	}
}
