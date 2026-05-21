// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package keychain

import (
	"regexp"
	"strings"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// nameRE enforces the macforge-<TEAM>-<PURPOSE>[.keychain-db] convention.
//
//	macforge-XYZ1234567-signing
//	macforge-XYZ1234567-signing.keychain-db
var nameRE = regexp.MustCompile(`^macforge-[A-Za-z0-9]{3,12}-[a-z0-9][a-z0-9-]{1,30}[a-z0-9](?:\.keychain-db)?$`)

// DefaultName returns the canonical keychain name for a team/purpose pair.
func DefaultName(team, purpose string) string {
	return "macforge-" + team + "-" + purpose
}

// ValidateName enforces the convention and refuses any login.* name.
func ValidateName(name string) error {
	if strings.HasPrefix(name, "login.") {
		return mferrors.NewKeychain(mferrors.CodeKeychainNonStandardName,
			"keychain.ValidateName",
			"login.keychain is forbidden by MacForge policy",
			mferrors.WithDetails(map[string]any{"name": name}))
	}
	if !nameRE.MatchString(name) {
		return mferrors.NewKeychain(mferrors.CodeKeychainNonStandardName,
			"keychain.ValidateName",
			"name does not match macforge-<TEAM>-<PURPOSE> convention",
			mferrors.WithHint("Use DefaultName(team, purpose) or set keychain.allow_nonstandard: true"),
			mferrors.WithDetails(map[string]any{"name": name}))
	}
	return nil
}
