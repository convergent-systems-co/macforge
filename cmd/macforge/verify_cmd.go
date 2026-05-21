// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
	"github.com/convergent-systems-co/macforge/internal/apple/spctl"
	"github.com/convergent-systems-co/macforge/internal/verify"
)

func newVerifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <path>",
		Short: "Verify codesign + spctl + Gatekeeper for the given artifact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("apple.verify", false)
			if err != nil {
				return err
			}
			r := newRunnerWithAudit(rt)
			svc := verify.New(codesign.New(r), spctl.New(r))

			at := assessTypeForPath(args[0])
			res, vErr := svc.Verify(cmd.Context(), args[0], at)
			return rt.emit(verifyResult{Result: res}, vErr)
		},
	}
}

func assessTypeForPath(path string) spctl.AssessType {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pkg":
		return spctl.AssessTypeInstall
	case ".dmg":
		return spctl.AssessTypeOpen
	default:
		return spctl.AssessTypeExec
	}
}

type verifyResult struct {
	Result verify.Result `json:"result"`
}

func (r verifyResult) SchemaName() string { return "macforge.v1.apple.verify" }
func (r verifyResult) HumanLines() []string {
	out := []string{"Path: " + r.Result.Path}
	if r.Result.Codesign.OK {
		out = append(out, "  ✓ codesign --verify passed")
	} else {
		out = append(out, "  ✗ codesign --verify FAILED")
	}
	if r.Result.Spctl.Accepted {
		out = append(out, "  ✓ spctl --assess accepted ("+r.Result.Spctl.Source+")")
	} else {
		out = append(out, "  ✗ spctl --assess REJECTED")
	}
	return out
}
