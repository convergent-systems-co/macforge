// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
	"github.com/convergent-systems-co/macforge/internal/apple/security"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
	"github.com/convergent-systems-co/macforge/internal/signing"
)

func newSignCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sign <path>",
		Short: "Sign one or more macOS artifacts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("sign", true)
			if err != nil {
				return err
			}
			r := newRunnerWithAudit(rt)
			svc := signing.New(codesign.New(r), security.New(r))

			result := signResult{Team: rt.cfg.Team}
			for _, artifact := range args {
				res, signErr := svc.Sign(cmd.Context(), rt.cfg, artifact)
				if signErr != nil {
					return rt.emit(result, signErr)
				}
				result.Signed = append(result.Signed, res)
			}
			if len(result.Signed) == 0 {
				return rt.emit(result, mferrors.NewSigning(mferrors.CodeSignVerificationFail,
					"sign.Run", "no artifacts signed"))
			}
			return rt.emit(result, nil)
		},
	}
}

type signResult struct {
	Team   string           `json:"team"`
	Signed []signing.Result `json:"signed"`
}

func (r signResult) SchemaName() string { return "macforge.v1.sign" }
func (r signResult) HumanLines() []string {
	if len(r.Signed) == 0 {
		return []string{"No artifacts signed"}
	}
	out := []string{"Team: " + r.Team}
	for _, s := range r.Signed {
		shortSHA := s.IdentitySHA1
		if len(shortSHA) > 12 {
			shortSHA = shortSHA[:12]
		}
		out = append(out, "  ✓ "+s.Path+"  ("+shortSHA+"…)")
	}
	return out
}
