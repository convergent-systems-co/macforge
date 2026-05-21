// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package config

import "github.com/convergent-systems-co/macforge/internal/keychain"

// ResolveKeychainName returns the keychain name every Apple verb should
// operate on, given a previously-validated *Config.
//
// Rules:
//
//   - If cfg.Keychain.Name is set, it is returned verbatim. The static
//     validator in Load (validate()) is responsible for ensuring Name is
//     well-formed (matches macforge-<TEAM>-<PURPOSE>) and consistent with
//     cfg.Team. When cfg.Keychain.AllowNonstandard is true the team-segment
//     consistency check is skipped by the validator; ResolveKeychainName
//     does not re-validate.
//   - Otherwise the team-derived default keychain.DefaultName(cfg.Team,
//     "signing") is returned.
//
// Callers MUST use this helper rather than reading cfg.Keychain.Name
// directly. The original bug closed by #13 was caused by signing reading
// cfg.Keychain.Name directly and getting a stale value that disagreed with
// cfg.Team; routing every read through this resolver makes it impossible to
// silently consume a not-yet-validated value.
func ResolveKeychainName(cfg *Config) string {
	if cfg == nil {
		return ""
	}
	if cfg.Keychain.Name != "" {
		return cfg.Keychain.Name
	}
	return keychain.DefaultName(cfg.Team, "signing")
}
