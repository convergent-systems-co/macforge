// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newIdentityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Manage Developer ID identities",
	}
	cmd.AddCommand(
		stubSub("create", "Create a new private key + CSR in the dedicated keychain"),
		stubSub("import", "Import Developer ID certificate(s) into the dedicated keychain"),
		stubSub("list", "List identities across known keychains"),
		stubSub("rotate", "Rotate the certificate; archive the old"),
		stubSub("status", "Show certificate validity, expiration, team"),
		stubSub("export", "Encrypted export for CI consumption"),
	)
	return cmd
}

func stubSub(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mferrors.NewConfig("MF-CONFIG-001", cmd.CommandPath(), "not yet implemented")
		},
	}
}
