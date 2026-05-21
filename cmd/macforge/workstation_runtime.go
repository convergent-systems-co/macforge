// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	workstationconfig "github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// workstationRuntimeFor builds a *workstationconfig.Runtime from cobra's
// resolved flag values for a workstation subcommand invocation. Inherited
// macforge globals (--dry-run, --verbose, --no-color) are read from gflags;
// workstation-specific flags (--workstation-repo, --quiet, --yes) are read
// from cmd.Flags() via the persistent flags on newWorkstationCmd.
func workstationRuntimeFor(cmd *cobra.Command) *workstationconfig.Runtime {
	repo, _ := cmd.Flags().GetString("workstation-repo")
	quiet, _ := cmd.Flags().GetBool("quiet")
	yes, _ := cmd.Flags().GetBool("yes")

	return &workstationconfig.Runtime{
		RepoPath: repo,
		DryRun:   gflags.dryRun,
		Verbose:  gflags.verbose,
		Quiet:    quiet,
		Yes:      yes,
		NoColor:  gflags.noColor,
	}
}
