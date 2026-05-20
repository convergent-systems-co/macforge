// Package config holds the runtime configuration parsed from CLI flags and
// (later) from ~/.config/macheim/config.yaml. Subcommands receive a *Runtime
// pointer at construction time; the root command's Before hook populates it
// from flag values before any Action runs.
package config

import (
	"errors"
	"fmt"
	"os"
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

// ResolveRepoPath returns the configured macheim repo path and the source
// that produced it ("flag" or "env"), or ("", "", nil) when no source is
// configured (the embed-fallback case from GOALS.md).
//
// Sub-issue #6 expands this to cover the remaining steps of the 5-step
// discovery chain: ~/.config/macheim/config.yaml, conventional paths
// (~/src/macheim, ~/code/macheim), and embed-fallback formalization. The
// signature is stable across that expansion — callers (e.g. doctor) do not
// change when #6 lands.
//
// rt.RepoPath already incorporates the cli.EnvVars("MACHEIM_REPO") value
// source set up in cmd/root.go, so in practice the os.Getenv fallback below
// is reached only when both the flag and cli's source chain miss the env.
// Defensive — keeps the resolver self-contained for tests that don't go
// through the cli flag-parsing path.
func (r *Runtime) ResolveRepoPath() (path, source string, err error) {
	if r.RepoPath != "" {
		return r.RepoPath, "flag", nil
	}
	if v := os.Getenv("MACHEIM_REPO"); v != "" {
		return v, "env", nil
	}
	return "", "", nil
}
