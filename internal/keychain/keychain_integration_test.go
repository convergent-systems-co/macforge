// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build integration

package keychain_test

import (
	"context"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/keychain"
)

func TestManager_List_OneIdentity(t *testing.T) {
	t.Setenv("MACFORGE_TEST_PW", "supersecret")
	r, err := apple.NewFakeRunner("../../testdata")
	if err != nil {
		t.Fatalf("NewFakeRunner: %v", err)
	}
	m := keychain.NewManager(security.New(r))

	ids, err := m.List(context.Background(), "macforge-XYZ-signing")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("len = %d, want 1", len(ids))
	}
}
