// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package keychain

import (
	"os"
	"strings"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// ResolveSecret turns a config reference (env:VAR or keyring:macforge:<id>)
// into a live secret value. Inline values are rejected here, mirroring the
// config-load validation; this is defense in depth.
//
// keyring: backend is left as a stub returning an error for v0.1 — the
// macOS keychain bridge is a follow-up.
func ResolveSecret(ref string) (string, error) {
	switch {
	case strings.HasPrefix(ref, "env:"):
		name := strings.TrimPrefix(ref, "env:")
		v := os.Getenv(name)
		if v == "" {
			return "", mferrors.NewKeychain(mferrors.CodeKeychainLocked,
				"keychain.ResolveSecret",
				"env var "+name+" is unset or empty")
		}
		return v, nil
	case strings.HasPrefix(ref, "keyring:"):
		return "", mferrors.NewKeychain(mferrors.CodeKeychainLocked,
			"keychain.ResolveSecret",
			"keyring: backend is not implemented in v0.1",
			mferrors.WithHint("Use env:VAR_NAME for now"))
	default:
		return "", mferrors.NewKeychain(mferrors.CodeKeychainLocked,
			"keychain.ResolveSecret",
			"secret reference must be 'env:VAR' or 'keyring:<id>'")
	}
}
