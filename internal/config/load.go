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
// defaults (./macforge.yaml for ProjectPath, the user-level config for UserPath).
type LoadOptions struct {
	ProjectPath string // explicit project config; default ./macforge.yaml
	UserPath    string // explicit user config; default $XDG/MacForge/config.yaml
}

// Load reads MacForge config with viper layering:
//
//	flag        (handled by caller before invoking Load)
//	env         (MACFORGE_* prefix)
//	project YAML (./macforge.yaml unless overridden)
//	user YAML   (~/Library/.../MacForge/config.yaml unless overridden)
//	defaults
//
// Returns a *mferrors.Error wrapping any read or validation failure.
func Load(opts LoadOptions) (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("MACFORGE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	projectPath := opts.ProjectPath
	if projectPath == "" {
		projectPath = "macforge.yaml"
	}

	if _, err := os.Stat(projectPath); err != nil {
		return nil, mferrors.NewConfig(mferrors.CodeConfigMissing, "config.Load",
			fmt.Sprintf("project config not found at %s", projectPath),
			mferrors.WithHint("Run `macforge init` to scaffold one, or pass --config <path>"),
			mferrors.WithCause(err),
		)
	}
	v.SetConfigFile(projectPath)
	if err := v.MergeInConfig(); err != nil {
		return nil, mferrors.NewConfig(mferrors.CodeConfigInvalid, "config.Load",
			fmt.Sprintf("failed to parse %s", projectPath),
			mferrors.WithCause(err),
		)
	}

	userPath := opts.UserPath
	if userPath == "" {
		userPath = filepath.Join(UserConfigDir(), "config.yaml")
	}
	if _, err := os.Stat(userPath); err == nil {
		v.SetConfigFile(userPath)
		_ = v.MergeInConfig() // user config is optional; ignore parse errors silently? No — surface.
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
