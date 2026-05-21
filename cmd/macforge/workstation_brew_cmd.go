// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/brew"
)

func newWorkstationBrewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brew",
		Short: "Homebrew operations",
	}
	cmd.AddCommand(
		newWorkstationBrewInstallCmd(),
		newWorkstationBrewBundleCmd(),
	)
	return cmd
}

func newWorkstationBrewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install Homebrew itself (arch-aware)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return brew.Install(rt)
		},
	}
}

func newWorkstationBrewBundleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bundle",
		Short: "Apply the repo Brewfile (brew bundle wrapper with embed fallback)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return brew.Apply(rt)
		},
	}
}
