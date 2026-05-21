// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package identity manages Developer ID identity lifecycle: import,
// list, and status. Key and CSR generation are planned for a later
// iteration (see GOALS.md "Identity" — "Private key generation" and
// "CSR generation" are out of scope for v0.1).
package identity

import (
	"context"

	"github.com/convergent-systems-co/macforge/internal/apple/security"
)

// Service is the identity orchestrator.
type Service struct {
	sec *security.Client
}

// New returns a Service driven by sec.
func New(sec *security.Client) *Service { return &Service{sec: sec} }

// List returns valid signing identities in the named keychain.
func (s *Service) List(ctx context.Context, keychain string) ([]security.Identity, error) {
	return s.sec.FindIdentities(ctx, keychain, "codesigning")
}
