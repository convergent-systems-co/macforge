// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package signing

import (
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple/security"
)

func TestSelectIdentity_ByTeam(t *testing.T) {
	ids := []security.Identity{
		{SHA1Fingerprint: "AAA", CommonName: "Developer ID Application: ACME (XYZ1234567)"},
		{SHA1Fingerprint: "BBB", CommonName: "Developer ID Application: Other (ABC9876543)"},
	}
	got, err := selectIdentityByTeam(ids, "XYZ1234567")
	if err != nil {
		t.Fatalf("selectIdentityByTeam: %v", err)
	}
	if got.SHA1Fingerprint != "AAA" {
		t.Fatalf("got %q, want AAA", got.SHA1Fingerprint)
	}
}

func TestSelectIdentity_NotFound(t *testing.T) {
	ids := []security.Identity{
		{SHA1Fingerprint: "BBB", CommonName: "Developer ID Application: Other (ABC9876543)"},
	}
	if _, err := selectIdentityByTeam(ids, "XYZ1234567"); err == nil {
		t.Fatal("expected error for missing team match")
	}
}

func TestSelectIdentity_EmptyList(t *testing.T) {
	if _, err := selectIdentityByTeam(nil, "XYZ1234567"); err == nil {
		t.Fatal("expected error for empty identity list")
	}
}
