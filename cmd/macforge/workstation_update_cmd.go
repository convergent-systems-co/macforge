// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/brew"
	"github.com/convergent-systems-co/macforge/internal/workstation/dotfiles"
)

func newWorkstationUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Reconcile drift between this Mac and the workstation repo",
	}
	cmd.AddCommand(
		newWorkstationUpdateLocalToRemoteCmd(),
		newWorkstationUpdateRemoteToLocalCmd(),
	)
	return cmd
}

func newWorkstationUpdateLocalToRemoteCmd() *cobra.Command {
	var module string
	cmd := &cobra.Command{
		Use:   "local-to-remote",
		Short: "Ratchet local drift back into the repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			switch module {
			case "brew":
				return brew.UpdateLocalToRemote(rt)
			case "dotfiles":
				// Result discarded: dotfiles.UpdateLocalToRemote renders output internally
				// via renderReverseDrift/logReverse. Hooking up a cmd-layer render path
				// here would double-print. See internal/workstation/dotfiles/local_to_remote.go.
				_, err := dotfiles.UpdateLocalToRemote(rt)
				return err
			case "all", "":
				if err := brew.UpdateLocalToRemote(rt); err != nil {
					return err
				}
				// Result discarded: dotfiles.UpdateLocalToRemote renders output internally
				// via renderReverseDrift/logReverse. Hooking up a cmd-layer render path
				// here would double-print. See internal/workstation/dotfiles/local_to_remote.go.
				_, err := dotfiles.UpdateLocalToRemote(rt)
				return err
			default:
				return cmd.Help()
			}
		},
	}
	cmd.Flags().StringVar(&module, "module", "all", "brew | dotfiles | all")
	return cmd
}

func newWorkstationUpdateRemoteToLocalCmd() *cobra.Command {
	var module string
	var prune, noPull bool
	cmd := &cobra.Command{
		Use:   "remote-to-local",
		Short: "Apply repo state to this Mac",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			switch module {
			case "brew":
				return brew.UpdateRemoteToLocal(rt, prune, noPull)
			case "all", "":
				return brew.UpdateRemoteToLocal(rt, prune, noPull)
			default:
				return cmd.Help()
			}
		},
	}
	cmd.Flags().StringVar(&module, "module", "all", "brew | all (dotfiles remote-to-local lands in a future commit)")
	cmd.Flags().BoolVar(&prune, "prune", false, "DESTRUCTIVE: uninstalls every formula/cask not in Brewfile (brew bundle --cleanup); use --dry-run first")
	cmd.Flags().BoolVar(&noPull, "no-pull", false, "skip git pull before applying")
	return cmd
}
