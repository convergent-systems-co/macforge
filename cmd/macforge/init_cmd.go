// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func newInitCmd() *cobra.Command {
	var team string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold the global macforge.yaml (~/.config/macforge/macforge.yaml)",
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

// runInit writes the GLOBAL scaffold to ~/.config/macforge/macforge.yaml
// (or the XDG override). Project-local override files are user-authored;
// init never writes one. See ADR-0015.
//
// The scaffold contains only the identity-shaped fields (team, identity,
// keychain). Artifact-shaped fields (sign.entitlements, package.formats,
// publish.github.repo) belong in project-local override files, not here.
func runInit(team string) (initResult, error) {
	path := macforgeYAMLPath()

	if _, err := os.Stat(path); err == nil {
		return initResult{}, mferrors.NewConfig(mferrors.CodeConfigInvalid, "init.Run",
			"macforge.yaml already exists at "+path,
			mferrors.WithHint("Delete it first if you really want to re-scaffold"))
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return initResult{}, mferrors.NewConfig(mferrors.CodeConfigInvalid, "init.Run",
			"failed to create parent directory for "+path,
			mferrors.WithCause(err))
	}

	scaffold := fmt.Sprintf(`# MacForge global config — identity-shaped fields only.
# See ADR-0015 for the project-vs-global distinction.
# Project-specific overrides (entitlements, package formats, publish.github.repo)
# go in ./macforge.yaml in the project root, not here.
version: 1
team: %s
identity:
  signing: developer-id-application
keychain:
  name: macforge-%s-signing
  unlock: env:MACFORGE_KEYCHAIN_PASSWORD
sign:
  hardened_runtime: true
  timestamp: true
notarize:
  asc_profile: macforge-prod
  wait: true
  staple: true
`, team, team)

	if err := os.WriteFile(path, []byte(scaffold), 0o644); err != nil {
		return initResult{}, mferrors.NewConfig(mferrors.CodeConfigInvalid, "init.Run",
			"failed to write "+path, mferrors.WithCause(err))
	}
	return initResult{Path: path, Team: team}, nil
}
