// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print MacForge version",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("macforge %s\n", version)
		},
	}
}
