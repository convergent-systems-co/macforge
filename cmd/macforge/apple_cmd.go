// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

// newAppleCmd returns the `apple` subtree holding every Apple-platform
// release operation MacForge ships today: keychain lifecycle, identity
// lifecycle, sign, verify, and the still-stubbed package/notarize/publish/
// release/audit verbs. See ADR-0017.
func newAppleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apple",
		Short: "Apple-platform release operations (signing, notarization, packaging, publish)",
		Long: `Apple-platform release operations: keychain and identity management,
signing, verification, packaging, notarization, and publication of macOS
software distributed outside the App Store.

This subtree contains everything MacForge does today. The peer subtree
` + "`macforge workstation`" + ` holds Mac workstation operations (Homebrew,
dotfiles, zsh, macOS defaults).`,
	}
	cmd.AddCommand(
		newInitCmd(),
		newConfigCmd(),
		newIdentityCmd(),
		newKeychainCmd(),
		newSignCmd(),
		newPackageCmd(),
		newNotarizeCmd(),
		newVerifyCmd(),
		newPublishCmd(),
		newReleaseCmd(),
		newAuditCmd(),
	)
	return cmd
}
