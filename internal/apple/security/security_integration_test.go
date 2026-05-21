// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build integration

package security_test

import (
	"context"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
)

func TestFindIdentities_Empty(t *testing.T) {
	r, err := apple.NewFakeRunner("../../../testdata")
	if err != nil {
		t.Fatalf("NewFakeRunner: %v", err)
	}
	c := security.New(r)

	ids, err := c.FindIdentities(context.Background(), "macforge-empty-signing", "codesigning")
	if err != nil {
		t.Fatalf("FindIdentities: %v", err)
	}
	if len(ids) != 0 {
		t.Fatalf("len = %d, want 0", len(ids))
	}
}

func TestFindIdentities_One(t *testing.T) {
	r, err := apple.NewFakeRunner("../../../testdata")
	if err != nil {
		t.Fatalf("NewFakeRunner: %v", err)
	}
	c := security.New(r)

	ids, err := c.FindIdentities(context.Background(), "macforge-XYZ-signing", "codesigning")
	if err != nil {
		t.Fatalf("FindIdentities: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("len = %d, want 1", len(ids))
	}
	want := "ABCDEF0123456789ABCDEF0123456789ABCDEF01"
	if ids[0].SHA1Fingerprint != want {
		t.Fatalf("SHA1 = %q, want %q", ids[0].SHA1Fingerprint, want)
	}
	if ids[0].CommonName != "Developer ID Application: ACME Inc. (XYZ1234567)" {
		t.Fatalf("CommonName = %q", ids[0].CommonName)
	}
}
