// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package audit

import "testing"

func TestRedactor_MasksSecrets(t *testing.T) {
	r := NewRedactor([]string{"sekret-password", "abc123"})
	in := "running with --password sekret-password and token=abc123"
	want := "running with --password [REDACTED] and token=[REDACTED]"
	if got := r.Apply(in); got != want {
		t.Fatalf("Apply() = %q, want %q", got, want)
	}
}

func TestRedactor_EmptySecretsListIsNoOp(t *testing.T) {
	r := NewRedactor(nil)
	in := "no secrets here"
	if got := r.Apply(in); got != in {
		t.Fatalf("Apply() with nil secrets changed input: got %q", got)
	}
}

func TestRedactor_IgnoresEmptyStrings(t *testing.T) {
	r := NewRedactor([]string{"", "real-secret"})
	in := "leak real-secret"
	want := "leak [REDACTED]"
	if got := r.Apply(in); got != want {
		t.Fatalf("Apply() = %q, want %q", got, want)
	}
}
