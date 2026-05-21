// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config

import (
	"os"
	"path/filepath"
)

// UserConfigDir returns the macOS standard user-level config root for MacForge.
// On non-darwin hosts (used in cross-platform tests and CI) it falls back to
// $HOME/.macforge-user.
func UserConfigDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return ".macforge-user"
	}
	if runtimeIsDarwin {
		return filepath.Join(home, "Library", "Application Support", "MacForge")
	}
	return filepath.Join(home, ".macforge-user")
}

// ProjectAuditDir returns the project-local audit directory (./.macforge/audit).
func ProjectAuditDir(cwd string) string {
	return filepath.Join(cwd, ".macforge", "audit")
}
