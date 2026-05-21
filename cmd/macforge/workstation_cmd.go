// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

// newWorkstationCmd returns the `workstation` subtree: Homebrew, dotfiles,
// zsh, macOS defaults, and the read-only doctor/status checks. Peer to
// `newAppleCmd()`. See ADR-0017 (namespace pattern) and ADR-0018
// (peer-name decision).
func newWorkstationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workstation",
		Short: "Mac workstation operations (Homebrew, dotfiles, zsh, macOS defaults)",
		Long: `Mac workstation operations: bootstrap a fresh Mac to a known-good state
defined in a git repo, reconcile drift in either direction, and run
read-only doctor/status checks.

Most mutating verbs (bootstrap, brew install/bundle, dotfiles apply,
zsh setup, macos defaults, downloads, update *) are currently stubbed —
they print "not implemented yet" and exit zero. The read-only verbs
(status, doctor) work today.

Originated as github.com/polliard/macheim; merged into macforge on
2026-05-21.`,
	}

	pflags := cmd.PersistentFlags()
	pflags.String("workstation-repo", "", "path to your workstation repo (overrides discovery; also reads MACHEIM_REPO — will be MACFORGE_WORKSTATION_REPO once viper layering wires through)")
	pflags.BoolP("quiet", "q", false, "suppress non-error output")
	pflags.BoolP("yes", "y", false, "skip confirmation prompts")

	cmd.AddCommand(
		newWorkstationBootstrapCmd(),
		newWorkstationBrewCmd(),
		newWorkstationDotfilesCmd(),
		newWorkstationZshCmd(),
		newWorkstationMacosCmd(),
		newWorkstationDownloadsCmd(),
		newWorkstationUpdateCmd(),
		newWorkstationStatusCmd(),
		newWorkstationDoctorCmd(),
	)
	return cmd
}

// notImplementedRunE returns a cobra RunE that prints the standard
// "not implemented yet (see issue #N)" message and exits zero. Mirrors
// the macforge stub idiom used by package/notarize/publish/release.
// issue is the GitHub issue number that tracks the real implementation.
func notImplementedRunE(verb string, issue int) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		_, _ = cmd.OutOrStdout().Write([]byte(verb + ": not implemented yet (see issue #" + itoaWS(issue) + ")\n"))
		return nil
	}
}

// itoaWS is a small int-to-decimal helper for the stub printer. Renamed
// from itoa to avoid colliding with any helper that may live alongside
// the apple cmd files (e.g., identity_cmd.go's intToStr).
func itoaWS(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	n := len(buf)
	for i > 0 {
		n--
		buf[n] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		n--
		buf[n] = '-'
	}
	return string(buf[n:])
}
