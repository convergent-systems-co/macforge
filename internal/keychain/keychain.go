// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package keychain manages dedicated MacForge keychains. Names follow the
// macforge-<TEAM>-<PURPOSE> convention; passwords resolve through
// ResolveSecret. The underlying tool is apple/security.
package keychain

import (
	"context"

	"github.com/convergent-systems-co/macforge/internal/apple/security"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// Manager is the high-level keychain orchestrator.
type Manager struct {
	sec *security.Client
}

// NewManager wires a Manager to a security.Client.
func NewManager(sec *security.Client) *Manager { return &Manager{sec: sec} }

// CreateOptions describes a Create call. AllowNonstandard skips the naming
// validator; SecretRef is the env:/keyring: reference.
type CreateOptions struct {
	Name             string
	SecretRef        string
	AllowNonstandard bool
	LockOnSleep      bool
	LockTimeoutSecs  int
}

// Create resolves the password, validates the name, calls security
// create-keychain, then applies set-keychain-settings.
func (m *Manager) Create(ctx context.Context, opts CreateOptions) error {
	if !opts.AllowNonstandard {
		if err := ValidateName(opts.Name); err != nil {
			return err
		}
	}
	pw, err := ResolveSecret(opts.SecretRef)
	if err != nil {
		return err
	}
	if err := m.sec.CreateKeychain(ctx, opts.Name, pw); err != nil {
		return err
	}
	if err := m.sec.SetSettings(ctx, opts.Name, opts.LockOnSleep, opts.LockTimeoutSecs); err != nil {
		return err
	}
	return nil
}

// Unlock resolves the password and unlocks the keychain.
func (m *Manager) Unlock(ctx context.Context, name, secretRef string) error {
	pw, err := ResolveSecret(secretRef)
	if err != nil {
		return err
	}
	return m.sec.UnlockKeychain(ctx, name, pw)
}

// Delete removes the keychain. Refuses login.* names (the security
// wrapper also refuses; this is defense in depth).
func (m *Manager) Delete(ctx context.Context, name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	return m.sec.DeleteKeychain(ctx, name)
}

// List enumerates identities under a keychain. Used by the CLI list verb.
func (m *Manager) List(ctx context.Context, name string) ([]security.Identity, error) {
	if err := ValidateName(name); err != nil {
		// Tolerate non-MacForge names in list with a warning-only flow —
		// for v0.1 we return the validation error so the CLI can decide.
		return nil, err
	}
	return m.sec.FindIdentities(ctx, name, "codesigning")
}

// guard for unused import if security types ever drift.
var _ = mferrors.ErrKeychain
