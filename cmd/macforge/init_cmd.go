// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newInitCmd() *cobra.Command {
	var team string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold macforge.yaml in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := newRuntime("init", false)
			if err != nil {
				return err
			}

			result, runErr := runInit(team)
			return rt.emit(result, runErr)
		},
	}
	cmd.Flags().StringVar(&team, "team", "", "Apple Developer Team ID (10 chars)")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

// initResult implements output.Outputter.
type initResult struct {
	Path string `json:"path"`
	Team string `json:"team"`
}

func (r initResult) SchemaName() string   { return "macforge.v1.init" }
func (r initResult) HumanLines() []string { return []string{"Wrote " + r.Path, "Team:  " + r.Team} }

func runInit(team string) (initResult, error) {
	path := macforgeYAMLPath()
	if _, err := os.Stat(path); err == nil {
		return initResult{}, mferrors.NewConfig(mferrors.CodeConfigInvalid, "init.Run",
			"macforge.yaml already exists at "+path,
			mferrors.WithHint("Delete it first if you really want to re-scaffold"))
	}

	scaffold := fmt.Sprintf(`version: 1
team: %s
identity:
  signing: developer-id-application
keychain:
  name: macforge-%s-signing
  unlock: env:MACFORGE_KEYCHAIN_PASSWORD
sign:
  hardened_runtime: true
  timestamp: true
  entitlements: ./Entitlements.plist
notarize:
  asc_profile: macforge-prod
  wait: true
  staple: true
package:
  formats:
    - zip
`, team, team)

	if err := os.WriteFile(path, []byte(scaffold), 0o644); err != nil {
		return initResult{}, mferrors.NewConfig(mferrors.CodeConfigInvalid, "init.Run",
			"failed to write macforge.yaml", mferrors.WithCause(err))
	}
	return initResult{Path: path, Team: team}, nil
}
