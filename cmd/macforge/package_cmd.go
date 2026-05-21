// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newPackageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "package <path>",
		Short: "Build .zip, .dmg, .pkg, or .app bundles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mferrors.NewConfig("MF-CONFIG-001", "package.Run", "not yet implemented")
		},
	}
}
