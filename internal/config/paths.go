// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config

import (
	"os"
	"path/filepath"
)

// UserConfigDir returns the directory holding the MacForge config.
//
// Honors $XDG_CONFIG_HOME if set, else falls back to $HOME/.config/macforge.
// Uniform across macOS and Linux per ADR-0015 — the path is not
// platform-specific.
func UserConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "macforge")
	}
	home, _ := os.UserHomeDir()
	if home == "" {
		return filepath.Join(".config", "macforge")
	}
	return filepath.Join(home, ".config", "macforge")
}

// ConfigPath returns the canonical macforge.yaml location.
func ConfigPath() string {
	return filepath.Join(UserConfigDir(), "macforge.yaml")
}

// AuditDir returns the user-home audit directory at ~/.macforge/audit.
// Per ADR-0016 the audit log is user-scoped (not per-project) and
// per-invocation (not daily-rotated). Honors $HOME; falls back to
// `.macforge/audit` if $HOME is unset (extremely unusual; cwd-relative).
func AuditDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return filepath.Join(".macforge", "audit")
	}
	return filepath.Join(home, ".macforge", "audit")
}
