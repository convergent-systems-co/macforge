// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package keychain

import "testing"

func TestResolveSecret_Env(t *testing.T) {
	t.Setenv("MACFORGE_TEST_PW", "topsecret")
	got, err := ResolveSecret("env:MACFORGE_TEST_PW")
	if err != nil {
		t.Fatalf("ResolveSecret: %v", err)
	}
	if got != "topsecret" {
		t.Fatalf("got %q, want topsecret", got)
	}
}

func TestResolveSecret_EnvMissing(t *testing.T) {
	if _, err := ResolveSecret("env:DEFINITELY_NOT_SET_XYZ"); err == nil {
		t.Fatal("expected error for unset env var")
	}
}

func TestResolveSecret_InlineRejected(t *testing.T) {
	if _, err := ResolveSecret("hunter2"); err == nil {
		t.Fatal("expected error for inline password reference")
	}
}
