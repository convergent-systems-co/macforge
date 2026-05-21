// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newSignCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sign <path>",
		Short: "Sign one or more macOS artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mferrors.NewConfig("MF-CONFIG-001", "sign.Run", "not yet implemented")
		},
	}
}
