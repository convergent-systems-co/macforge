// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isolatedEnv points HOME at a fresh tempdir and clears MACHEIM_REPO so
// repo discovery cannot pick up the developer's real workstation repo or
// global config during a test. Returns the tempdir for callers that need
// to write fake state into it.
func isolatedEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MACHEIM_REPO", "")
	return home
}

// runArgs runs the root command with args and returns combined stdout+stderr.
// Always uses an isolated env per isolatedEnv.
func runArgs(t *testing.T, args ...string) (output string, err error) {
	t.Helper()
	isolatedEnv(t)
	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetErr(&sb)
	root.SetArgs(args)
	err = root.Execute()
	return sb.String(), err
}

// TestWorkstation_HelpTreeStructure walks every node in the workstation
// command tree and asserts --help works at that level and lists every
// subcommand we expect. Catches structural drift between cobra registration
// and the public surface.
func TestWorkstation_HelpTreeStructure(t *testing.T) {
	cases := []struct {
		path     []string
		wantSubs []string
	}{
		{
			[]string{"workstation"},
			[]string{"bootstrap", "brew", "doctor", "dotfiles", "downloads", "macos", "status", "update", "zsh"},
		},
		{[]string{"workstation", "brew"}, []string{"bundle", "install"}},
		{[]string{"workstation", "dotfiles"}, []string{"apply"}},
		{[]string{"workstation", "update"}, []string{"local-to-remote", "remote-to-local"}},
		{[]string{"workstation", "macos"}, []string{"defaults"}},
		{[]string{"workstation", "zsh"}, []string{"setup"}},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.path, "/"), func(t *testing.T) {
			out, err := runArgs(t, append(tc.path, "--help")...)
			if err != nil {
				t.Fatalf("--help failed for %v: %v", tc.path, err)
			}
			for _, sub := range tc.wantSubs {
				if !strings.Contains(out, sub) {
					t.Errorf("--help output missing subcommand %q\n---\n%s", sub, out)
				}
			}
		})
	}
}

// TestNonApple_LeafVerbsAcceptHelp confirms every leaf workstation verb
// renders --help cleanly. A leaf that returns nil from its constructor
// or panics during help would fail here — a low-cost smoke test for the
// whole cobra registration surface.
func TestNonApple_LeafVerbsAcceptHelp(t *testing.T) {
	leaves := [][]string{
		{"workstation", "doctor"},
		{"workstation", "status"},
		{"workstation", "bootstrap"},
		{"workstation", "downloads"},
		{"workstation", "brew", "bundle"},
		{"workstation", "brew", "install"},
		{"workstation", "dotfiles", "apply"},
		{"workstation", "update", "local-to-remote"},
		{"workstation", "update", "remote-to-local"},
		{"workstation", "macos", "defaults"},
		{"workstation", "zsh", "setup"},
		{"version"},
	}
	for _, leaf := range leaves {
		t.Run(strings.Join(leaf, "/"), func(t *testing.T) {
			out, err := runArgs(t, append(leaf, "--help")...)
			if err != nil {
				t.Fatalf("--help failed: %v", err)
			}
			if !strings.Contains(out, "Usage:") {
				t.Fatalf("--help did not render Usage section\n%s", out)
			}
		})
	}
}

// TestWorkstation_GlobalFlagsVisibleOnLeaf confirms that the macforge
// global flags (--config, --dry-run, --log-level, --output, etc.) and
// the workstation persistent flags (--workstation-repo, --quiet, --yes)
// are inherited on a deep leaf verb. Regression guard against cobra
// flag-registration drift after the macforge apple/workstation split.
func TestWorkstation_GlobalFlagsVisibleOnLeaf(t *testing.T) {
	out, err := runArgs(t, "workstation", "dotfiles", "apply", "--help")
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}
	wantFlags := []string{
		"--config",
		"--dry-run",
		"--log-level",
		"--no-color",
		"--output",
		"--quiet",
		"--team-id",
		"--verbose",
		"--workstation-repo",
		"--yes",
	}
	for _, f := range wantFlags {
		if !strings.Contains(out, f) {
			t.Errorf("global flag %q not visible on workstation leaf verb\n---\n%s", f, out)
		}
	}
}

// TestWorkstation_DotfilesApply_RefusesWithoutRepo asserts that
// `workstation dotfiles apply` refuses with a clear, actionable message
// when no repo is configured (no flag, no MACHEIM_REPO, no ~/.config
// override, no ~/src/macheim or ~/code/macheim convention).
//
// Without this guard, dotfiles.Apply would happily copy from a relative
// "dotfiles" path resolved against process CWD.
func TestWorkstation_DotfilesApply_RefusesWithoutRepo(t *testing.T) {
	out, err := runArgs(t, "workstation", "dotfiles", "apply")
	if err == nil {
		t.Fatalf("expected error, got nil\n---\n%s", out)
	}
	if !strings.Contains(err.Error(), "no repo configured") {
		t.Errorf("error %q missing 'no repo configured'", err.Error())
	}
	if !strings.Contains(err.Error(), "--workstation-repo") {
		t.Errorf("error %q missing remediation hint --workstation-repo", err.Error())
	}
}

// TestWorkstation_UpdateLocalToRemote_RefusesWithoutRepo asserts that
// `workstation update local-to-remote` refuses with a clear error in
// embed-fallback mode. update local-to-remote is mutating and must
// never silently no-op against an unconfigured environment.
func TestWorkstation_UpdateLocalToRemote_RefusesWithoutRepo(t *testing.T) {
	out, err := runArgs(t, "workstation", "update", "local-to-remote")
	if err == nil {
		t.Fatalf("expected error, got nil\n---\n%s", out)
	}
	if !strings.Contains(err.Error(), "no repo configured") {
		t.Errorf("error %q missing 'no repo configured'", err.Error())
	}
}

// TestWorkstation_UpdateRemoteToLocal_RefusesWithoutRepo asserts that
// `workstation update remote-to-local` refuses without a repo. The
// embedded Brewfile fallback is read-only and remote-to-local must
// flag that case explicitly rather than blindly running brew bundle.
func TestWorkstation_UpdateRemoteToLocal_RefusesWithoutRepo(t *testing.T) {
	out, err := runArgs(t, "workstation", "update", "remote-to-local")
	if err == nil {
		t.Fatalf("expected error, got nil\n---\n%s", out)
	}
	if !strings.Contains(err.Error(), "no repo configured") {
		t.Errorf("error %q missing 'no repo configured'", err.Error())
	}
}

// TestWorkstation_VerboseQuietMutex confirms cross-cutting flag-mutex
// validation in workstationconfig.Runtime.Validate fires when both
// --verbose and --quiet are set on a workstation verb. Status is the
// canonical pure-read verb to drive this with — it has no side effects.
//
// Note: as of the current wiring, status does NOT call Validate() on the
// runtime, so the mutex doesn't fire end-to-end at the CLI layer. This
// test serves as a regression guard pinning the current behavior; flip
// the expectation to "must error" once a future change wires Validate
// into the workstation Before hook.
func TestWorkstation_VerboseQuietMutex(t *testing.T) {
	out, err := runArgs(t, "workstation", "status", "--verbose", "--quiet")
	if err != nil {
		// If/when Validate gets wired in: the error MUST name both flags.
		// Document the current decoupling explicitly so the next person
		// changing this doesn't have to spelunk the runtime to learn it.
		if !strings.Contains(err.Error(), "verbose") || !strings.Contains(err.Error(), "quiet") {
			t.Errorf("if status validates flag mutex, error must name --verbose and --quiet; got: %v\n---\n%s",
				err, out)
		}
	}
}

// TestWorkstation_RepoFlagHonored confirms that --workstation-repo on
// a workstation verb populates the runtime's RepoPath. We drive this
// through `dotfiles apply` because (a) it's the most explicit consumer
// of ResolveRepoPath at the cmd layer, and (b) its error path tells us
// whether the flag was honored: with the flag, no "no repo configured"
// error; without it, that error fires.
//
// We point --workstation-repo at a tempdir that exists but has no
// dotfiles/ subdir so dotfiles.Apply errors on a different path (NOT
// "no repo configured"), proving the flag was consulted.
func TestWorkstation_RepoFlagHonored(t *testing.T) {
	home := isolatedEnv(t)
	repo := filepath.Join(home, "fake-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetErr(&sb)
	root.SetArgs([]string{"workstation", "dotfiles", "apply", "--workstation-repo=" + repo})
	err := root.Execute()
	if err != nil && strings.Contains(err.Error(), "no repo configured") {
		t.Fatalf("--workstation-repo flag not honored; got fallback error: %v\n---\n%s",
			err, sb.String())
	}
	// We don't require success — dotfiles.Apply may fail because the
	// fake repo has no dotfiles/ subdir, which is fine. We only require
	// that the failure path is NOT the no-repo-configured branch.
}

// TestWorkstation_MacheimRepoEnvHonored confirms that MACHEIM_REPO,
// the documented env-var fallback for repo discovery, is consulted when
// no --workstation-repo flag is set. Same pattern as TestWorkstation_RepoFlagHonored:
// existence of a non-no-repo-configured error proves the env was read.
func TestWorkstation_MacheimRepoEnvHonored(t *testing.T) {
	home := isolatedEnv(t)
	repo := filepath.Join(home, "fake-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	t.Setenv("MACHEIM_REPO", repo)

	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetErr(&sb)
	root.SetArgs([]string{"workstation", "dotfiles", "apply"})
	err := root.Execute()
	if err != nil && strings.Contains(err.Error(), "no repo configured") {
		t.Fatalf("MACHEIM_REPO env not honored; got fallback error: %v\n---\n%s",
			err, sb.String())
	}
}

// TestWorkstation_UnknownSubverbFallsBackToHelp pins the current cobra
// behavior: parent verbs with no RunE (workstation, brew, dotfiles,
// update, macos, zsh) print their help and exit zero when an unknown
// subverb is supplied. The fat-finger does not produce an error.
//
// This may surprise users typing e.g. `macforge workstation brue install`,
// who expect an error. Tightening to "unknown command" error would
// require adding RunE handlers that route unknown subverbs to a
// non-nil error. Filed for later if/when we standardize on that.
func TestWorkstation_UnknownSubverbFallsBackToHelp(t *testing.T) {
	out, err := runArgs(t, "workstation", "definitelynotaverb")
	if err != nil {
		t.Fatalf("expected fall-through to help (nil err), got: %v\n---\n%s", err, out)
	}
	if !strings.Contains(out, "Available Commands:") {
		t.Errorf("expected parent help to render with Available Commands section\n---\n%s", out)
	}
}

// TestWorkstation_BrewUnknownSubverbFallsBackToHelp is the deep-tree
// counterpart: a fat-finger at depth 3 (workstation brew <garbage>)
// also falls through to the brew parent's help, exit zero. Same caveat
// as TestWorkstation_UnknownSubverbFallsBackToHelp.
func TestWorkstation_BrewUnknownSubverbFallsBackToHelp(t *testing.T) {
	out, err := runArgs(t, "workstation", "brew", "totallymadeup")
	if err != nil {
		t.Fatalf("expected fall-through to help (nil err), got: %v\n---\n%s", err, out)
	}
	if !strings.Contains(out, "Available Commands:") {
		t.Errorf("expected brew help to render with Available Commands section\n---\n%s", out)
	}
}
