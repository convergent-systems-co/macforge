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
	"github.com/convergent-systems-co/macforge/internal/keychain"
)

func requireEnv(t *testing.T, names ...string) {
	t.Helper()
	for _, n := range names {
		if os.Getenv(n) == "" {
			t.Skipf("skipping e2e: %s is unset", n)
		}
	}
}

func TestE2E_KeychainLifecycle(t *testing.T) {
	requireEnv(t, "MACFORGE_E2E_TEAM", "MACFORGE_E2E_KEYCHAIN_PASSWORD")

	team := os.Getenv("MACFORGE_E2E_TEAM")
	t.Setenv("MACFORGE_TEST_PW", os.Getenv("MACFORGE_E2E_KEYCHAIN_PASSWORD"))

	r := apple.NewExecRunner(nil)
	m := keychain.NewManager(security.New(r))

	name := keychain.DefaultName(team, "e2e")
	ctx := context.Background()

	// Cleanup: delete first in case a previous run leaked. Ignore errors.
	_ = m.Delete(ctx, name)

	if err := m.Create(ctx, keychain.CreateOptions{
		Name:            name,
		SecretRef:       "env:MACFORGE_TEST_PW",
		LockOnSleep:     true,
		LockTimeoutSecs: 3600,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() { _ = m.Delete(context.Background(), name) })

	if err := m.Unlock(ctx, name, "env:MACFORGE_TEST_PW"); err != nil {
		t.Fatalf("Unlock: %v", err)
	}

	ids, err := m.List(ctx, name)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// Fresh keychain has no identities.
	if len(ids) != 0 {
		t.Fatalf("fresh keychain has %d identities, want 0", len(ids))
	}
}
