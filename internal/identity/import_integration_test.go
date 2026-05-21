// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build integration

package identity_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/identity"
)

func TestService_Import_Success(t *testing.T) {
	dir := t.TempDir()
	stub := filepath.Join(dir, "test-cert.pem")
	if err := os.WriteFile(stub, []byte("dummy"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// FakeRunner matches on argv only, not file content — the stub path is
	// what the args reference.

	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	r, err := apple.NewFakeRunner(filepath.Join(wd, "..", "..", "testdata"))
	if err != nil {
		t.Fatalf("NewFakeRunner: %v", err)
	}
	svc := identity.New(security.New(r))

	err = svc.Import(context.Background(), identity.ImportOptions{
		File:     "./test-cert.pem",
		Keychain: "macforge-XYZ-signing",
	})
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
}
