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

// IdentityConfig holds the Developer ID identity kinds MacForge will use
// (signing certificate type, optional installer certificate type).
type IdentityConfig struct {
	Signing   string `mapstructure:"signing" yaml:"signing"`
	Installer string `mapstructure:"installer" yaml:"installer"`
}

// KeychainConfig holds the dedicated keychain's name, unlock-secret
// reference, lock policy, and validator escape-hatch.
type KeychainConfig struct {
	Name             string `mapstructure:"name" yaml:"name"`
	Unlock           string `mapstructure:"unlock" yaml:"unlock"`
	AllowNonstandard bool   `mapstructure:"allow_nonstandard" yaml:"allow_nonstandard"`
	LockTimeoutSecs  int    `mapstructure:"lock_timeout" yaml:"lock_timeout"`
	PersistUnlock    bool   `mapstructure:"persist_unlock" yaml:"persist_unlock"`
}

// SignConfig holds the per-codesign options: Hardened Runtime, secure
// timestamp, and entitlements plist path.
type SignConfig struct {
	HardenedRuntime bool   `mapstructure:"hardened_runtime" yaml:"hardened_runtime"`
	Timestamp       bool   `mapstructure:"timestamp" yaml:"timestamp"`
	Entitlements    string `mapstructure:"entitlements" yaml:"entitlements"`
}

// NotarizeConfig holds the notarytool profile name, wait-for-completion
// preference, and stapling toggle.
type NotarizeConfig struct {
	ASCProfile string `mapstructure:"asc_profile" yaml:"asc_profile"`
	Wait       bool   `mapstructure:"wait" yaml:"wait"`
	Staple     bool   `mapstructure:"staple" yaml:"staple"`
}

// PackageConfig holds the desired output artifact formats (zip, dmg, pkg, app).
type PackageConfig struct {
	Formats []string `mapstructure:"formats" yaml:"formats"`
}

// PublishConfig holds destinations for the publish verb. Currently only
// GitHub Releases; more destinations land here as they're implemented.
type PublishConfig struct {
	GitHub GitHubConfig `mapstructure:"github" yaml:"github"`
}

// GitHubConfig describes a GitHub Releases publish target: the
// owner/repo coordinate and whether to ship a draft release first.
type GitHubConfig struct {
	Repo  string `mapstructure:"repo" yaml:"repo"`
	Draft bool   `mapstructure:"draft" yaml:"draft"`
}

// AuditConfig configures the JSONL audit writer's user-level mirror and
// whether to include full stdout/stderr bodies (default: hashes only).
type AuditConfig struct {
	UserMirror bool `mapstructure:"user_mirror" yaml:"user_mirror"`
	Bodies     bool `mapstructure:"bodies" yaml:"bodies"`
}
