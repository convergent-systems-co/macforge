// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Scaffold macforge.yaml in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mferrors.NewConfig("MF-CONFIG-001", "init.Run", "not yet implemented")
		},
	}
}
