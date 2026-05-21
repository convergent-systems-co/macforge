// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationMacosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "macos",
		Short: "macOS operations",
	}
	cmd.AddCommand(newWorkstationMacosDefaultsCmd())
	return cmd
}

func newWorkstationMacosDefaultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "defaults",
		Short: "Apply the repo macOS defaults manifest (planned)",
		RunE:  notImplementedRunE("macos defaults", 19),
	}
}
