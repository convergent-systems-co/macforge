// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/config"
)

func TestLoad_ProjectYAML(t *testing.T) {
	dir := t.TempDir()
	yaml := `
version: 1
team: XYZ1234567
identity:
  signing: developer-id-application
keychain:
  name: macforge-XYZ1234567-signing
  unlock: env:MACFORGE_KEYCHAIN_PASSWORD
sign:
  hardened_runtime: true
  timestamp: true
  entitlements: ./Entitlements.plist
`
	path := filepath.Join(dir, "macforge.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cfg, err := config.Load(config.LoadOptions{ProjectPath: path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Version != 1 {
		t.Fatalf("Version = %d, want 1", cfg.Version)
	}
	if cfg.Team != "XYZ1234567" {
		t.Fatalf("Team = %q, want XYZ1234567", cfg.Team)
	}
	if cfg.Keychain.Name != "macforge-XYZ1234567-signing" {
		t.Fatalf("Keychain.Name = %q", cfg.Keychain.Name)
	}
	if !cfg.Sign.HardenedRuntime {
		t.Fatal("Sign.HardenedRuntime = false, want true")
	}
}

func TestLoad_EnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	yaml := `
version: 1
team: XYZ1234567
keychain:
  name: from-yaml
`
	path := filepath.Join(dir, "macforge.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	t.Setenv("MACFORGE_TEAM", "ABC9876543")

	cfg, err := config.Load(config.LoadOptions{ProjectPath: path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Team != "ABC9876543" {
		t.Fatalf("env override failed: Team = %q, want ABC9876543", cfg.Team)
	}
}

func TestLoad_RejectsInlinePassword(t *testing.T) {
	dir := t.TempDir()
	yaml := `
version: 1
team: XYZ1234567
keychain:
  name: macforge-XYZ1234567-signing
  unlock: hunter2
`
	path := filepath.Join(dir, "macforge.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := config.Load(config.LoadOptions{ProjectPath: path})
	if err == nil {
		t.Fatal("expected error for inline password, got nil")
	}
}

func TestLoad_MissingFileIsError(t *testing.T) {
	_, err := config.Load(config.LoadOptions{ProjectPath: "/nonexistent/macforge.yaml"})
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}
