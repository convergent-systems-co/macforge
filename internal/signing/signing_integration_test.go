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

// TestService_Sign_UsesResolvedKeychainName_FromEmptyName proves the #13
// fix: when cfg.Keychain.Name is unset, signing.Sign asks the security CLI
// for identities in the team-derived default name (macforge-<TEAM>-signing),
// not in an empty string.
//
// The fixture pair used:
//   - testdata/security/find-identity-one-derived   — find-identity for
//     keychain `macforge-XYZ1234567-signing` (the resolver-derived name).
//   - testdata/codesign/sign-derived-keychain       — codesign --sign with
//     the same keychain string.
//
// If the resolver were not wired (i.e., Sign still read cfg.Keychain.Name
// directly), the FakeRunner would error out with "no fixture matches" on
// the find-identity call and this test would fail.
func TestService_Sign_UsesResolvedKeychainName_FromEmptyName(t *testing.T) {
	r, _ := apple.NewFakeRunner("../../testdata")
	svc := signing.New(codesign.New(r), security.New(r))

	cfg := &config.Config{
		Version: 1,
		Team:    "XYZ1234567",
		// Keychain.Name intentionally unset — resolver must derive
		// macforge-XYZ1234567-signing.
		Sign: config.SignConfig{
			HardenedRuntime: true,
			Timestamp:       true,
			Entitlements:    "./Entitlements.plist",
		},
	}

	res, err := svc.Sign(context.Background(), cfg, "./build/MyApp.app")
	if err != nil {
		t.Fatalf("Sign with derived keychain name: %v", err)
	}
	if res.IdentitySHA1 == "" {
		t.Fatal("IdentitySHA1 empty — FakeRunner did not match the find-identity-one-derived fixture; signing did not resolve to macforge-XYZ1234567-signing")
	}
}

// TestService_Sign_UsesResolvedKeychainName_FromExplicitName proves that
// when cfg.Keychain.Name IS set, Sign honors it verbatim (the resolver
// returns the explicit name) rather than overriding it with the derived
// default. Both branches of ResolveKeychainName must be exercised.
//
// The existing fixtures (find-identity-one + sign-signing-svc-success) are
// keyed on `macforge-XYZ-signing`, so we set Keychain.Name to that
// explicit value and Team to XYZ1234567 (which is what the fixture's
// identity CN contains).
func TestService_Sign_UsesResolvedKeychainName_FromExplicitName(t *testing.T) {
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
		t.Fatalf("Sign with explicit keychain name: %v", err)
	}
	if res.IdentitySHA1 == "" {
		t.Fatal("IdentitySHA1 empty")
	}
}
