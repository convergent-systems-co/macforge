// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build integration

package signing_test

import (
	"context"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/config"
	"github.com/convergent-systems-co/macforge/internal/signing"
)

func TestService_Sign_RoundTrip(t *testing.T) {
	r, _ := apple.NewFakeRunner("../../testdata")
	svc := signing.New(codesign.New(r), security.New(r))

	cfg := &config.Config{
		Version: 1,
		Team:    "XYZ1234567",
		Keychain: config.KeychainConfig{
			Name: "macforge-XYZ-signing",
		},
		Sign: config.SignConfig{
			HardenedRuntime: true,
			Timestamp:       true,
			Entitlements:    "./Entitlements.plist",
		},
	}

	res, err := svc.Sign(context.Background(), cfg, "./build/MyApp.app")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if res.IdentitySHA1 == "" {
		t.Fatal("IdentitySHA1 empty")
	}
}
