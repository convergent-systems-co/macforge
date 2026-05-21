// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/dotfiles"
)

func newWorkstationDotfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotfiles",
		Short: "Dotfile operations",
	}
	cmd.AddCommand(newWorkstationDotfilesApplyCmd())
	return cmd
}

func newWorkstationDotfilesApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Copy <repo>/dotfiles/ into $HOME with backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			repoPath, _, err := rt.ResolveRepoPath()
			if err != nil {
				return err
			}
			homePath, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			_, err = dotfiles.Apply(rt, repoPath, homePath)
			return err
		},
	}
}
