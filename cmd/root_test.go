package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/polliard/macheim/internal/config"
)

// runRoot constructs a fresh Runtime + root command, redirects output to
// in-memory buffers, and runs the given args. Returns stdout, stderr, the
// populated Runtime, and the run error.
func runRoot(t *testing.T, args ...string) (string, string, *config.Runtime, error) {
	t.Helper()
	rt := &config.Runtime{
		Version:   "v0.0.1",
		Commit:    "deadbeef",
		BuildDate: "2026-01-01T00:00:00Z",
	}
	root := NewRoot(rt)
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	root.Writer = out
	root.ErrWriter = errOut
	err := root.Run(context.Background(), append([]string{"macheim"}, args...))
	return out.String(), errOut.String(), rt, err
}

func TestRoot_HelpShowsGlobalFlags(t *testing.T) {
	t.Parallel()
	stdout, _, _, err := runRoot(t, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"--repo", "--dry-run", "--verbose", "--quiet", "--yes", "--no-color"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help output missing %q\nfull output:\n%s", want, stdout)
		}
	}
}

func TestRoot_VersionShowsBuildIdentity(t *testing.T) {
	t.Parallel()
	stdout, _, _, err := runRoot(t, "--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "v0.0.1 (commit deadbeef, built 2026-01-01T00:00:00Z)"
	if !strings.Contains(stdout, want) {
		t.Errorf("--version output missing %q\nfull output:\n%s", want, stdout)
	}
}

func TestRoot_HelpListsEveryCommand(t *testing.T) {
	t.Parallel()
	stdout, _, _, err := runRoot(t, "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{
		"bootstrap", "brew", "zsh", "dotfiles", "macos",
		"downloads", "update", "status", "doctor",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("help missing command %q\nfull:\n%s", want, stdout)
		}
	}
}

// doctor is no longer a stub; integration coverage moved to doctor_test.go
// (TestDoctor_RunsAndProducesSummary).

func TestStub_BrewBundle(t *testing.T) {
	t.Parallel()
	stdout, _, _, err := runRoot(t, "brew", "bundle")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "brew bundle: not implemented yet (see issue #13)\n"
	if stdout != want {
		t.Errorf("stdout:\n  got:  %q\n  want: %q", stdout, want)
	}
}

func TestStub_UpdateLocalToRemote(t *testing.T) {
	t.Parallel()
	stdout, _, _, err := runRoot(t, "update", "local-to-remote")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "update local-to-remote: not implemented yet (see issue #15)\n"
	if stdout != want {
		t.Errorf("stdout:\n  got:  %q\n  want: %q", stdout, want)
	}
}

func TestStub_ZshSetup(t *testing.T) {
	t.Parallel()
	stdout, _, _, err := runRoot(t, "zsh", "setup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "zsh setup: not implemented yet (see issue #20)\n"
	if stdout != want {
		t.Errorf("stdout:\n  got:  %q\n  want: %q", stdout, want)
	}
}

func TestRoot_FlagValidation_VerboseAndQuiet(t *testing.T) {
	t.Parallel()
	// Use a stub subcommand (bootstrap) rather than doctor: doctor's real checks
	// invoke cli.Exit("", 1) on failure, which terminates the test process via
	// the framework's default OsExiter. On Linux CI, doctor's xcode/brew probes
	// fail and the whole package's tests collateral-die. Stub commands print
	// and return nil — Before still fires, validation error still surfaces.
	_, _, _, err := runRoot(t, "--verbose", "--quiet", "bootstrap")
	if err == nil {
		t.Fatal("expected error for --verbose --quiet together, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention mutually exclusive; got %q", err.Error())
	}
}

func TestRoot_FlagInheritance_DryRunSetInRuntime(t *testing.T) {
	t.Parallel()
	// See TestRoot_FlagValidation_VerboseAndQuiet for why bootstrap (a stub) is
	// used instead of doctor.
	_, _, rt, err := runRoot(t, "--dry-run", "bootstrap")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rt.DryRun {
		t.Error("--dry-run did not propagate into rt.DryRun")
	}
}
