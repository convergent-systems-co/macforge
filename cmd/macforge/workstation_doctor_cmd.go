// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/doctor"
)

func newWorkstationDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Sanity-check the workstation environment",
		Long: `Read-only diagnostic. Verifies xcode-select, Homebrew, repo discovery,
config directory, and the shell rc file. Exits zero on all-pass; exits
one with per-check remediation hints when something is broken.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return doctor.Run(rt, cmd.OutOrStdout())
		},
	}
}
