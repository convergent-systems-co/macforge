// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package keychain

import "testing"

func TestValidateName_OK(t *testing.T) {
	cases := []string{
		"macforge-XYZ1234567-signing",
		"macforge-ABC9876543-dev",
		"macforge-XYZ1234567-staging",
	}
	for _, name := range cases {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", name, err)
		}
	}
}

func TestValidateName_Bad(t *testing.T) {
	cases := []string{
		"login.keychain",
		"login.keychain-db",
		"random-name",
		"macforge-",
		"macforge--signing",
		"macforge-XYZ-",
	}
	for _, name := range cases {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", name)
		}
	}
}

func TestDefaultName(t *testing.T) {
	got := DefaultName("XYZ1234567", "signing")
	want := "macforge-XYZ1234567-signing"
	if got != want {
		t.Fatalf("DefaultName = %q, want %q", got, want)
	}
}
