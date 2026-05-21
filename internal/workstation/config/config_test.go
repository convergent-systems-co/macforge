//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeConfig writes a config.yaml under the given HOME with the given body.
// Returns the absolute path of the written file.
func writeConfig(t *testing.T, home, body string) string {
	t.Helper()
	dir := filepath.Join(home, ".config", "macheim")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func TestConfig_Load_Missing(t *testing.T) {
	// Tests in this file share env state via t.Setenv; do not parallelize.
	t.Setenv("HOME", t.TempDir())

	got, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != (Config{}) {
		t.Fatalf("want zero Config, got %#v", got)
	}
}

func TestConfig_Load_Present(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeConfig(t, home, "repo_path: /tmp/macheim\n")

	got, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.RepoPath != "/tmp/macheim" {
		t.Fatalf("repo_path: got %q, want %q", got.RepoPath, "/tmp/macheim")
	}
}

func TestConfig_Load_Malformed(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Unbalanced bracket — yaml.v3 rejects this.
	writeConfig(t, home, "repo_path: [unterminated\n")

	_, err := Load()
	if err == nil {
		t.Fatalf("want non-nil error for malformed yaml, got nil")
	}
}