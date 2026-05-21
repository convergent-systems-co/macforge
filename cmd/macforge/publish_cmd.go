// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newPublishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish <path>",
		Short: "Upload signed artifacts to GitHub Releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mferrors.NewConfig("MF-CONFIG-001", "publish.Run", "not yet implemented")
		},
	}
}
