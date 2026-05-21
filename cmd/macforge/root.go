// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"
)

// globalFlags holds inherited flag values; populated by cobra at run time.
type globalFlags struct {
	configPath string
	output     string
	logLevel   string
	teamID     string
	dryRun     bool
	noColor    bool
	verbose    bool
}

var gflags globalFlags

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "macforge",
		Short: "Civilization-grade Apple release infrastructure",
		Long: `MacForge provides deterministic, auditable, repeatable Apple signing,
certificate lifecycle management, packaging, notarization, verification,
and publishing for macOS software distributed outside the App Store.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pflags := root.PersistentFlags()
	pflags.StringVar(&gflags.configPath, "config", "", "path to global macforge.yaml (default: ~/.config/macforge/macforge.yaml)")
	pflags.StringVar(&gflags.output, "output", "", "output mode: human | json (default: human, auto json under GITHUB_ACTIONS)")
	pflags.StringVar(&gflags.logLevel, "log-level", "info", "error | warn | info | debug | trace")
	pflags.StringVar(&gflags.teamID, "team-id", "", "override config team selection")
	pflags.BoolVar(&gflags.dryRun, "dry-run", false, "show planned Apple invocations; emit zero")
	pflags.BoolVar(&gflags.noColor, "no-color", false, "disable ANSI output (also honors NO_COLOR)")
	pflags.BoolVarP(&gflags.verbose, "verbose", "v", false, "shortcut for --log-level=debug")

	// All Apple-platform release operations live under `macforge apple <verb>`.
	// `version` stays at the root (universal). Workstation operations (Homebrew,
	// dotfiles, zsh, macOS defaults) live under `macforge workstation <verb>` as
	// a peer. See docs/adr/0017-apple-command-namespace.md (namespace pattern)
	// and docs/adr/0018-peer-subtree-named-workstation.md (peer-name decision).
	root.AddCommand(
		newAppleCmd(),
		newVersionCmd(),
	)

	return root
}
