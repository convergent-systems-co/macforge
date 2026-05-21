// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

// Package verify composes codesign --verify and spctl --assess.
package verify

import (
	"context"

	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
	"github.com/convergent-systems-co/macforge/internal/apple/spctl"
)

// Service is the verification orchestrator.
type Service struct {
	cs *codesign.Client
	sc *spctl.Client
}

// New returns a Service.
func New(cs *codesign.Client, sc *spctl.Client) *Service {
	return &Service{cs: cs, sc: sc}
}

// Result aggregates codesign and spctl outcomes.
type Result struct {
	Path     string                `json:"path"`
	Codesign codesign.VerifyResult `json:"codesign"`
	Spctl    spctl.Result          `json:"spctl"`
	OK       bool                  `json:"ok"`
}

// Verify runs codesign --verify --strict and spctl --assess. The first
// failure short-circuits; the typed Result preserves both observations
// when both ran.
func (s *Service) Verify(ctx context.Context, path string, at spctl.AssessType) (Result, error) {
	out := Result{Path: path}

	csRes, csErr := s.cs.Verify(ctx, path)
	out.Codesign = csRes
	if csErr != nil {
		return out, csErr
	}

	scRes, scErr := s.sc.Assess(ctx, path, at)
	out.Spctl = scRes
	if scErr != nil {
		return out, scErr
	}

	out.OK = csRes.OK && scRes.Accepted
	return out, nil
}
