// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build darwin && e2e

package e2e

import (
	"context"
	"os"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/identity"
	"github.com/convergent-systems-co/macforge/internal/keychain"
)

func TestE2E_ImportAndList(t *testing.T) {
	requireEnv(t,
		"MACFORGE_E2E_TEAM",
		"MACFORGE_E2E_KEYCHAIN_PASSWORD",
		"MACFORGE_E2E_CERT_PATH", // path to a Developer ID Application .cer or .p12
	)

	team := os.Getenv("MACFORGE_E2E_TEAM")
	t.Setenv("MACFORGE_TEST_PW", os.Getenv("MACFORGE_E2E_KEYCHAIN_PASSWORD"))

	r := apple.NewExecRunner(nil)
	sec := security.New(r)
	m := keychain.NewManager(sec)
	svc := identity.New(sec)
	ctx := context.Background()
	name := keychain.DefaultName(team, "e2e-import")

	_ = m.Delete(ctx, name)
	if err := m.Create(ctx, keychain.CreateOptions{
		Name: name, SecretRef: "env:MACFORGE_TEST_PW", LockOnSleep: false, LockTimeoutSecs: 600,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = m.Delete(context.Background(), name) })

	if err := m.Unlock(ctx, name, "env:MACFORGE_TEST_PW"); err != nil {
		t.Fatalf("Unlock: %v", err)
	}

	err := svc.Import(ctx, identity.ImportOptions{
		File:        os.Getenv("MACFORGE_E2E_CERT_PATH"),
		Keychain:    name,
		P12Password: os.Getenv("MACFORGE_E2E_P12_PASSWORD"), // optional, "" for .cer/.pem
	})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	ids, err := svc.List(ctx, name)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("List returned 0 identities after Import")
	}
}
