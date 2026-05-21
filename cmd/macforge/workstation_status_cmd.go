// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/status"
)

func newWorkstationStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Read-only summary of drift between this Mac and the workstation repo",
		Long: `Read-only. Prints what's drifted between this Mac and the repo across
brew, repo, dotfiles, and macOS-defaults sections. Tolerates an
unconfigured environment — missing Homebrew, missing repo, etc. all
render as known states rather than errors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return status.Run(rt, cmd.OutOrStdout())
		},
	}
}
