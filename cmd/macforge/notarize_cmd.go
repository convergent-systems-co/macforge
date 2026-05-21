// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newNotarizeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notarize <path>",
		Short: "Submit, wait, and staple via notarytool",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mferrors.NewConfig("MF-CONFIG-001", "notarize.Run", "not yet implemented")
		},
	}
}
