// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"errors"
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
			// Guard the embed-fallback case: an empty repoPath means no
			// workstation repo is configured. Without this check,
			// dotfiles.Apply would call filepath.Join("", "dotfiles") which
			// resolves to "dotfiles" relative to process CWD — potentially
			// copying from an unintended source directory or silently no-oping.
			if repoPath == "" {
				return errors.New("dotfiles apply: no repo configured; clone the workstation repo and set --workstation-repo or MACHEIM_REPO first")
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
