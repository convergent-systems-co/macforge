// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// LoadOptions controls where Load reads from. Empty fields fall back to
// defaults: GlobalPath defaults to ${XDG_CONFIG_HOME:-$HOME/.config}/macforge/macforge.yaml,
// ProjectPath defaults to <cwd>/macforge.yaml.
type LoadOptions struct {
	GlobalPath  string // explicit global config; default: ConfigPath()
	ProjectPath string // explicit project override; default: ./macforge.yaml
}

// Load reads the MacForge config with viper layering:
//
//	flag        (handled by caller before invoking Load)
//	env         (MACFORGE_* prefix; dots in keys become underscores)
//	./macforge.yaml         (project override, OPTIONAL — only if present)
//	~/.config/macforge/macforge.yaml  (global base, REQUIRED)
//	defaults
//
// Per ADR-0015. The global file is required; without it Load returns
// MF-CONFIG-002 with a hint to run `macforge init`. The project-local
// file is OPTIONAL — its absence is not an error.
//
// Returns a *mferrors.Error wrapping any read or validation failure.
func Load(opts LoadOptions) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("MACFORGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	globalPath := opts.GlobalPath
	if globalPath == "" {
		globalPath = ConfigPath()
	}

	// Global config is required.
	if _, err := os.Stat(globalPath); err != nil {
		return nil, mferrors.NewConfig(mferrors.CodeConfigMissing, "config.Load",
			fmt.Sprintf("global config not found at %s", globalPath),
			mferrors.WithHint("Run `macforge init --team <TEAM>` to scaffold it"),
			mferrors.WithCause(err),
		)
	}
	v.SetConfigFile(globalPath)
	if err := v.MergeInConfig(); err != nil {
		return nil, mferrors.NewConfig(mferrors.CodeConfigInvalid, "config.Load",
			fmt.Sprintf("failed to parse global config %s", globalPath),
			mferrors.WithCause(err),
		)
	}

	// Project-local override is optional. If present, it merges on top of
	// global; per ADR-0015 it should carry only project-specific overrides
	// (entitlements, package formats, publish target) — but the loader doesn't
	// enforce that constraint in v0.1.
	projectPath := opts.ProjectPath
	if projectPath == "" {
		cwd, _ := os.Getwd()
		projectPath = filepath.Join(cwd, "macforge.yaml")
	}
	if _, err := os.Stat(projectPath); err == nil {
		v.SetConfigFile(projectPath)
		if err := v.MergeInConfig(); err != nil {
			return nil, mferrors.NewConfig(mferrors.CodeConfigInvalid, "config.Load",
				fmt.Sprintf("failed to parse project override %s", projectPath),
				mferrors.WithCause(err),
			)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, mferrors.NewConfig(mferrors.CodeConfigInvalid, "config.Load",
			"failed to unmarshal config", mferrors.WithCause(err))
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("version", 1)
	v.SetDefault("keychain.lock_timeout", 3600)
	v.SetDefault("sign.hardened_runtime", true)
	v.SetDefault("sign.timestamp", true)
	v.SetDefault("notarize.wait", true)
	v.SetDefault("notarize.staple", true)
	v.SetDefault("package.formats", []string{"zip"})
}

func validate(cfg *Config) error {
	if cfg.Version != 1 {
		return mferrors.NewConfig(mferrors.CodeConfigInvalid, "config.validate",
			fmt.Sprintf("unsupported config version %d (want 1)", cfg.Version))
	}
	if cfg.Keychain.Unlock != "" {
		if !strings.HasPrefix(cfg.Keychain.Unlock, "env:") && !strings.HasPrefix(cfg.Keychain.Unlock, "keyring:") {
			return mferrors.NewConfig(mferrors.CodeConfigInvalid, "config.validate",
				"keychain.unlock must be 'env:VAR' or 'keyring:macforge:<id>', not an inline value",
				mferrors.WithHint("Move the password to an environment variable and set keychain.unlock: env:VAR_NAME"))
		}
	}
	return nil
}
