// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config

// Config is the typed shape of macforge.yaml v1.
//
// Schema is versioned by the top-level Version field. Migrations between
// versions are explicit; never silently re-interpret values.
type Config struct {
	Version  int            `mapstructure:"version" yaml:"version"`
	Team     string         `mapstructure:"team" yaml:"team"`
	Identity IdentityConfig `mapstructure:"identity" yaml:"identity"`
	Keychain KeychainConfig `mapstructure:"keychain" yaml:"keychain"`
	Sign     SignConfig     `mapstructure:"sign" yaml:"sign"`
	Notarize NotarizeConfig `mapstructure:"notarize" yaml:"notarize"`
	Package  PackageConfig  `mapstructure:"package" yaml:"package"`
	Publish  PublishConfig  `mapstructure:"publish" yaml:"publish"`
	Audit    AuditConfig    `mapstructure:"audit" yaml:"audit"`
}

type IdentityConfig struct {
	Signing   string `mapstructure:"signing" yaml:"signing"`
	Installer string `mapstructure:"installer" yaml:"installer"`
}

type KeychainConfig struct {
	Name             string `mapstructure:"name" yaml:"name"`
	Unlock           string `mapstructure:"unlock" yaml:"unlock"`
	AllowNonstandard bool   `mapstructure:"allow_nonstandard" yaml:"allow_nonstandard"`
	LockTimeoutSecs  int    `mapstructure:"lock_timeout" yaml:"lock_timeout"`
	PersistUnlock    bool   `mapstructure:"persist_unlock" yaml:"persist_unlock"`
}

type SignConfig struct {
	HardenedRuntime bool   `mapstructure:"hardened_runtime" yaml:"hardened_runtime"`
	Timestamp       bool   `mapstructure:"timestamp" yaml:"timestamp"`
	Entitlements    string `mapstructure:"entitlements" yaml:"entitlements"`
}

type NotarizeConfig struct {
	ASCProfile string `mapstructure:"asc_profile" yaml:"asc_profile"`
	Wait       bool   `mapstructure:"wait" yaml:"wait"`
	Staple     bool   `mapstructure:"staple" yaml:"staple"`
}

type PackageConfig struct {
	Formats []string `mapstructure:"formats" yaml:"formats"`
}

type PublishConfig struct {
	GitHub GitHubConfig `mapstructure:"github" yaml:"github"`
}

type GitHubConfig struct {
	Repo  string `mapstructure:"repo" yaml:"repo"`
	Draft bool   `mapstructure:"draft" yaml:"draft"`
}

type AuditConfig struct {
	UserMirror bool `mapstructure:"user_mirror" yaml:"user_mirror"`
	Bodies     bool `mapstructure:"bodies" yaml:"bodies"`
}
