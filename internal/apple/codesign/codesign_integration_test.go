// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build integration

package codesign_test

import (
	"context"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
)

func TestSign_Success(t *testing.T) {
	r, _ := apple.NewFakeRunner("../../../testdata")
	c := codesign.New(r)

	err := c.Sign(context.Background(), codesign.SignOptions{
		IdentityCN:      "Developer ID Application: ACME (XYZ1234567)",
		Keychain:        "macforge-XYZ1234567-signing",
		HardenedRuntime: true,
		Timestamp:       true,
		Entitlements:    "./Entitlements.plist",
		Path:            "./build/MyApp.app",
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
}

func TestVerify_Success(t *testing.T) {
	r, _ := apple.NewFakeRunner("../../../testdata")
	c := codesign.New(r)

	res, err := c.Verify(context.Background(), "./build/MyApp.app")
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.OK {
		t.Fatal("VerifyResult.OK = false, want true")
	}
}
