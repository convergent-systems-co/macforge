// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationDownloadsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "downloads",
		Short: "Fetch optional downloads listed in the repo (planned)",
		RunE:  notImplementedRunE("downloads", 20),
	}
}
