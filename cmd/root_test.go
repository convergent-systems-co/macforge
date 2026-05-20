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

// Flag-validation integration (--verbose with --quiet returning an error
// from Before) requires a subcommand to actually invoke Before; that test
// lands in the commit that registers the doctor stub.
