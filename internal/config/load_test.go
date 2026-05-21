// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/config"
)

func writeYAML(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestLoad_GlobalOnly(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, `
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
`)

	cfg, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "nonexistent.yaml"),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Team != "XYZ1234567" {
		t.Fatalf("Team = %q, want XYZ1234567", cfg.Team)
	}
	if !cfg.Sign.HardenedRuntime {
		t.Fatal("Sign.HardenedRuntime = false, want true")
	}
}

func TestLoad_ProjectOverridesGlobal(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	project := filepath.Join(dir, "macforge.yaml")
	writeYAML(t, global, `
version: 1
team: GLOBAL_TEAM
sign:
  hardened_runtime: true
  entitlements: ./Global.plist
package:
  formats: [zip]
`)
	writeYAML(t, project, `
sign:
  entitlements: ./ProjectSpecific.plist
package:
  formats: [zip, dmg]
`)

	cfg, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: project,
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Team comes from global (not overridden).
	if cfg.Team != "GLOBAL_TEAM" {
		t.Fatalf("Team = %q, want GLOBAL_TEAM (project didn't override)", cfg.Team)
	}
	// Entitlements overridden by project.
	if cfg.Sign.Entitlements != "./ProjectSpecific.plist" {
		t.Fatalf("Sign.Entitlements = %q, want ./ProjectSpecific.plist", cfg.Sign.Entitlements)
	}
	// HardenedRuntime stays from global (project didn't touch it).
	if !cfg.Sign.HardenedRuntime {
		t.Fatal("Sign.HardenedRuntime = false, want true (carried from global)")
	}
	// Formats overridden by project.
	if len(cfg.Package.Formats) != 2 || cfg.Package.Formats[1] != "dmg" {
		t.Fatalf("Package.Formats = %v, want [zip dmg]", cfg.Package.Formats)
	}
}

func TestLoad_EnvOverridesBoth(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	project := filepath.Join(dir, "macforge.yaml")
	writeYAML(t, global, "version: 1\nteam: GLOBAL\n")
	writeYAML(t, project, "team: PROJECT\n")

	t.Setenv("MACFORGE_TEAM", "ENV_WINS")

	cfg, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: project,
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Team != "ENV_WINS" {
		t.Fatalf("Team = %q, want ENV_WINS", cfg.Team)
	}
}

func TestLoad_RejectsInlinePassword(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, `
version: 1
team: XYZ1234567
keychain:
  unlock: hunter2
`)

	_, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "nonexistent.yaml"),
	})
	if err == nil {
		t.Fatal("expected error for inline password, got nil")
	}
}

func TestLoad_MissingGlobalIsError(t *testing.T) {
	_, err := config.Load(config.LoadOptions{
		GlobalPath:  "/nonexistent/global.yaml",
		ProjectPath: "/nonexistent/project.yaml",
	})
	if err == nil {
		t.Fatal("expected error for missing global config, got nil")
	}
}

func TestLoad_MissingProjectIsOK(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, "version: 1\nteam: XYZ\n")

	_, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "no-such-file.yaml"),
	})
	if err != nil {
		t.Fatalf("missing project override should be silent: %v", err)
	}
}

func TestLoad_DefaultGlobalHonorsXDG(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	global := filepath.Join(dir, "macforge", "macforge.yaml")
	writeYAML(t, global, "version: 1\nteam: ZZZ0000000\n")

	// Set project path to a non-existent file in another tempdir so the
	// cwd default doesn't accidentally pick up a real macforge.yaml.
	cfg, err := config.Load(config.LoadOptions{
		ProjectPath: filepath.Join(t.TempDir(), "no-project.yaml"),
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Team != "ZZZ0000000" {
		t.Fatalf("Team = %q, want ZZZ0000000 (XDG-defaulted global)", cfg.Team)
	}
}

func TestLoad_RejectsMissingTeam(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, `
version: 1
`)
	_, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "no-project.yaml"),
	})
	if err == nil {
		t.Fatal("expected error for missing team, got nil")
	}
	if !strings.Contains(err.Error(), "team") {
		t.Fatalf("error should mention team: %v", err)
	}
}

func TestLoad_RejectsBadKeychainNameRegex(t *testing.T) {
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, `
version: 1
team: XYZ1234567
keychain:
  name: definitely-not-a-macforge-name
`)
	_, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "no-project.yaml"),
	})
	if err == nil {
		t.Fatal("expected error for malformed keychain.name, got nil")
	}
}

func TestLoad_RejectsTeamMismatch(t *testing.T) {
	// The #13 bug: keychain.name says one team, top-level team says another.
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, `
version: 1
team: XYZ1234567
keychain:
  name: macforge-ABC9876543-signing
`)
	_, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "no-project.yaml"),
	})
	if err == nil {
		t.Fatal("expected error for team/keychain.name mismatch, got nil")
	}
	// Hint should mention the team-segment vs cfg.team disagreement.
	if !strings.Contains(err.Error(), "XYZ1234567") && !strings.Contains(err.Error(), "ABC9876543") {
		t.Fatalf("error message should name the conflicting team values: %v", err)
	}
}

func TestLoad_AllowsNonstandardWhenOptedIn(t *testing.T) {
	// Escape hatch: allow_nonstandard: true → skip the team-consistency check.
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, `
version: 1
team: XYZ1234567
keychain:
  name: my-totally-custom-keychain-name
  allow_nonstandard: true
`)
	cfg, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "no-project.yaml"),
	})
	if err != nil {
		t.Fatalf("allow_nonstandard: true should let any keychain.name pass: %v", err)
	}
	if cfg.Keychain.Name != "my-totally-custom-keychain-name" {
		t.Fatalf("Keychain.Name = %q, want my-totally-custom-keychain-name", cfg.Keychain.Name)
	}
}

func TestLoad_AcceptsValidKeychainName(t *testing.T) {
	// Sanity check: the strict validator must NOT reject a well-formed
	// canonical name whose team segment matches.
	dir := t.TempDir()
	global := filepath.Join(dir, "global.yaml")
	writeYAML(t, global, `
version: 1
team: XYZ1234567
keychain:
  name: macforge-XYZ1234567-signing
  unlock: env:MACFORGE_KEYCHAIN_PASSWORD
`)
	cfg, err := config.Load(config.LoadOptions{
		GlobalPath:  global,
		ProjectPath: filepath.Join(dir, "no-project.yaml"),
	})
	if err != nil {
		t.Fatalf("canonical, matching keychain.name must be accepted: %v", err)
	}
	if cfg.Team != "XYZ1234567" || cfg.Keychain.Name != "macforge-XYZ1234567-signing" {
		t.Fatalf("config not preserved: team=%q name=%q", cfg.Team, cfg.Keychain.Name)
	}
}

func TestUserConfigDir_XDGOverridesHome(t *testing.T) {
	// Use t.TempDir so the path uses the platform-native separator without
	// hard-coding `/tmp/` (which on Windows resolves under C:\ and is
	// written with backslashes).
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	got := config.UserConfigDir()
	want := filepath.Join(xdg, "macforge")
	if got != want {
		t.Fatalf("UserConfigDir = %q, want %q", got, want)
	}
}

func TestUserConfigDir_FallsBackToHomeDotConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("UserHomeDir: %v", err)
	}
	got := config.UserConfigDir()
	want := filepath.Join(home, ".config", "macforge")
	if got != want {
		t.Fatalf("UserConfigDir = %q, want %q", got, want)
	}
}

func TestConfigPath_EndsInMacforgeYAML(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)
	got := config.ConfigPath()
	want := filepath.Join(xdg, "macforge", "macforge.yaml")
	if got != want {
		t.Fatalf("ConfigPath = %q, want %q", got, want)
	}
}
