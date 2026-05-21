// Package config holds the runtime configuration parsed from CLI flags and
// (later) from ~/.config/macheim/config.yaml. Subcommands receive a *Runtime
// pointer at construction time; the root command's Before hook populates it
// from flag values before any Action runs.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Runtime captures the inherited global flags and the build-time identity.
// All fields are zero-valued on construction; the root command's Before hook
// writes them.
//
// Output helpers added in a later sub-issue MUST consult NoColor before
// emitting ANSI. Until that helper lands no command emits ANSI, so the flag
// is stored-but-unused — which satisfies issue #5's "--no-color suppresses
// all ANSI globally" AC trivially.
type Runtime struct {
	RepoPath string
	DryRun   bool
	Verbose  bool
	Quiet    bool
	Yes      bool
	NoColor  bool

	Version   string
	Commit    string
	BuildDate string
}

// Validate enforces flag-interaction rules. Called from the root command's
// Before hook after the flags have been parsed.
func (r *Runtime) Validate() error {
	if r.Verbose && r.Quiet {
		return errors.New("--verbose and --quiet are mutually exclusive")
	}
	return nil
}

// VersionString formats the build identity for --version output.
func (r *Runtime) VersionString() string {
	return fmt.Sprintf("%s (commit %s, built %s)", r.Version, r.Commit, r.BuildDate)
}

// ResolveRepoPath walks the 5-step repo discovery chain from GOALS.md and
// returns the resolved path with the source that produced it. The source
// strings are stable identifiers callers can render in user output:
//
//	"flag"            — --repo / r.RepoPath was set
//	"env"             — MACHEIM_REPO was set
//	"config"          — ~/.config/macheim/config.yaml repo_path was set
//	"convention:src"  — ~/src/macheim exists as a directory
//	"convention:code" — ~/code/macheim exists as a directory
//
// Returning ("", "", nil) means no source resolved — the embed-fallback case.
// Doctor renders this as "no repo configured; running in embed-fallback
// mode" and future `update local-to-remote` refuses to mutate (see
// IsReadOnly).
//
// Errors from Load() other than "not exist" propagate so a misconfigured
// config file fails loudly rather than falling through to a convention path.
//
// rt.RepoPath already incorporates the cli.EnvVars("MACHEIM_REPO") value
// source set up in cmd/root.go, so in practice the os.Getenv fallback below
// is reached only when both the flag and cli's source chain miss the env.
// Defensive — keeps the resolver self-contained for tests that don't go
// through the cli flag-parsing path.
func (r *Runtime) ResolveRepoPath() (path, source string, err error) {
	// 1. --repo flag (already populated from cli.EnvVars by root's Before).
	if r.RepoPath != "" {
		return r.RepoPath, "flag", nil
	}
	// 2. MACHEIM_REPO env var (defensive — see note above).
	if v := os.Getenv("MACHEIM_REPO"); v != "" {
		return v, "env", nil
	}
	// 3. ~/.config/macheim/config.yaml repo_path.
	cfg, err := Load()
	if err != nil {
		return "", "", err
	}
	if cfg.RepoPath != "" {
		return cfg.RepoPath, "config", nil
	}
	// 4. Convention: ~/src/macheim then ~/code/macheim.
	home := homeDir()
	if p, ok := dirAt(filepath.Join(home, "src", "macheim")); ok {
		return p, "convention:src", nil
	}
	if p, ok := dirAt(filepath.Join(home, "code", "macheim")); ok {
		return p, "convention:code", nil
	}
	// 5. Embed-fallback.
	return "", "", nil
}

// IsReadOnly reports whether the runtime resolved to the embed-fallback
// case — no repo discovered. Mutating subcommands (notably
// `update local-to-remote`, landing in a later sub-issue) MUST refuse when
// this returns true, with a message directing the user to clone the repo
// first.
//
// Errors from ResolveRepoPath are treated as not-read-only so the caller's
// own error path runs unhindered — IsReadOnly is a precondition check, not a
// diagnostic.
func (r *Runtime) IsReadOnly() bool {
	path, source, err := r.ResolveRepoPath()
	if err != nil {
		return false
	}
	return path == "" && source == ""
}

// dirAt returns (path, true) when path exists and is a directory. Any other
// outcome — missing, file, permission denied — is reported as (_, false) so
// the caller can keep walking the discovery chain.
func dirAt(path string) (string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	if !info.IsDir() {
		return "", false
	}
	return path, true
}
