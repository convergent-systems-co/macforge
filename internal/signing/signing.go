// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package signing orchestrates artifact signing on top of apple/codesign.
package signing

import (
	"context"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	"github.com/convergent-systems-co/macforge/internal/config"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// Service is the signing orchestrator.
type Service struct {
	cs  *codesign.Client
	sec *security.Client
}

// New returns a Service driven by cs and sec.
func New(cs *codesign.Client, sec *security.Client) *Service {
	return &Service{cs: cs, sec: sec}
}

// Result is the typed outcome of a Sign call.
type Result struct {
	Path            string `json:"path"`
	IdentitySHA1    string `json:"identity_sha1"`
	IdentityName    string `json:"identity_name"`
	Team            string `json:"team"`
	HardenedRuntime bool   `json:"hardened_runtime"`
	Timestamp       bool   `json:"timestamp"`
}

// Sign locates the signing identity for cfg.Team in the resolved keychain
// (config.ResolveKeychainName), then runs codesign with the configured
// options. See #13: callers MUST go through ResolveKeychainName so the
// keychain name never disagrees with cfg.Team between Load and Sign.
func (s *Service) Sign(ctx context.Context, cfg *config.Config, artifact string) (Result, error) {
	keychainName := config.ResolveKeychainName(cfg)
	ids, err := s.sec.FindIdentities(ctx, keychainName, "codesigning")
	if err != nil {
		return Result{}, err
	}
	id, err := selectIdentityByTeam(ids, cfg.Team)
	if err != nil {
		return Result{}, err
	}

	opts := codesign.SignOptions{
		IdentityCN:      id.CommonName,
		Keychain:        keychainName,
		HardenedRuntime: cfg.Sign.HardenedRuntime,
		Timestamp:       cfg.Sign.Timestamp,
		Entitlements:    cfg.Sign.Entitlements,
		Path:            artifact,
	}
	if err := s.cs.Sign(ctx, opts); err != nil {
		return Result{}, err
	}

	return Result{
		Path:            artifact,
		IdentitySHA1:    id.SHA1Fingerprint,
		IdentityName:    id.CommonName,
		Team:            cfg.Team,
		HardenedRuntime: cfg.Sign.HardenedRuntime,
		Timestamp:       cfg.Sign.Timestamp,
	}, nil
}

func selectIdentityByTeam(ids []security.Identity, team string) (security.Identity, error) {
	if len(ids) == 0 {
		return security.Identity{}, mferrors.NewSigning(mferrors.CodeSignNoIdentity,
			"signing.selectIdentityByTeam",
			"no signing identities in the configured keychain",
			mferrors.WithHint("Run `macforge identity import --file <path>` first"))
	}
	needle := "(" + team + ")"
	for _, id := range ids {
		if strings.Contains(id.CommonName, needle) {
			return id, nil
		}
	}
	return security.Identity{}, mferrors.NewSigning(mferrors.CodeSignNoIdentity,
		"signing.selectIdentityByTeam",
		"no identity matches team "+team,
		mferrors.WithDetails(map[string]any{"team": team, "identity_count": len(ids)}))
}
