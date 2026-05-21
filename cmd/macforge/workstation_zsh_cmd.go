// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationZshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "zsh operations",
	}
	cmd.AddCommand(newWorkstationZshSetupCmd())
	return cmd
}

func newWorkstationZshSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure zsh as the macforge-managed shell (planned)",
		RunE:  notImplementedRunE("zsh setup", 18),
	}
}
