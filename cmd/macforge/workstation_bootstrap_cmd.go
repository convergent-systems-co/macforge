// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationBootstrapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "End-to-end fresh-Mac setup (planned)",
		RunE:  notImplementedRunE("bootstrap", 4),
	}
}
