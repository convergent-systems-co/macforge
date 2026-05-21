package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config mirrors the on-disk schema for ~/.config/macheim/config.yaml.
//
// The file is optional. When absent, Load returns a zero Config and a nil
// error so callers can treat "no file" the same as "no values set" and keep
// walking the discovery chain in ResolveRepoPath.
type Config struct {
	RepoPath string `yaml:"repo_path"`
}

// Load reads the user-level macheim config from $HOME/.config/macheim/
// config.yaml. Returns:
//   - (Config{}, nil)        when the file does not exist
//   - (Config{...}, nil)     when the file exists and parses cleanly
//   - (Config{}, non-nil)    when the file exists but cannot be read or parsed
//
// Permissions / open errors other than "not exist" surface as errors so a
// misconfigured environment fails loudly rather than silently falling through
// to a convention path.
func Load() (Config, error) {
	path := configPath()
	data, err := os.ReadFile(path) //nolint:gosec // path is built from a fixed layout under $HOME
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return c, nil
}

// configPath returns the conventional path for the user-level macheim
// config file. Centralized so tests can reason about a single layout.
func configPath() string {
	return filepath.Join(homeDir(), ".config", "macheim", "config.yaml")
}

// homeDir wraps os.UserHomeDir. UserHomeDir only errors when HOME is unset on
// Unix-likes, which is a fatal environment problem — macheim is a per-user
// macOS CLI that has no meaningful behavior without a home directory. Panic
// so the failure is loud and immediate rather than producing a path rooted at
// the empty string.
func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("resolve home directory: %w", err))
	}
	return h
}
