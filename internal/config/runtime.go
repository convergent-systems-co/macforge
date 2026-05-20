// Package config holds the runtime configuration parsed from CLI flags and
// (later) from ~/.config/macheim/config.yaml. Subcommands receive a *Runtime
// pointer at construction time; the root command's Before hook populates it
// from flag values before any Action runs.
package config

import (
	"errors"
	"fmt"
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
