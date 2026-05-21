// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newKeychainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keychain",
		Short: "Manage dedicated MacForge keychains",
	}
	cmd.AddCommand(
		stubSub("create", "Create a new managed keychain with the default name"),
		stubSub("delete", "Delete a managed keychain (refuses login.keychain)"),
		stubSub("list", "List MacForge-managed keychains"),
		stubSub("unlock", "Unlock the configured keychain for the current shell session"),
	)
	return cmd
}
