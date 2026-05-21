package status

import (
	"bytes"
	"strings"
	"testing"

	"github.com/polliard/macheim/internal/config"
)

// TestRun_QuietHidesDriftPlaceholders verifies the orchestrator's --quiet
// filter: Hidden:true rows (the three drift placeholders) are dropped
// when rt.Quiet is set, but brew + repo rows still print.
func TestRun_QuietHidesDriftPlaceholders(t *testing.T) {
	// Cannot t.Parallel — depends on HOME/MACHEIM_REPO + machine state. We
	// just assert presence/absence of substrings, not exact content.
	t.Setenv("MACHEIM_REPO", "")
	t.Setenv("HOME", t.TempDir())
	buf := &bytes.Buffer{}
	rt := &config.Runtime{Quiet: true}
	if err := Run(rt, buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "brew") {
		t.Errorf("--quiet should still print brew row; got %q", got)
	}
	if !strings.Contains(got, "repo") {
		t.Errorf("--quiet should still print repo row; got %q", got)
	}
	if strings.Contains(got, "drift:") {
		t.Errorf("--quiet should hide drift placeholders; got %q", got)
	}
}

// TestRun_DefaultShowsDriftPlaceholders verifies that without --quiet,
// all sections (including the three drift "?" placeholders) print.
func TestRun_DefaultShowsDriftPlaceholders(t *testing.T) {
	t.Setenv("MACHEIM_REPO", "")
	t.Setenv("HOME", t.TempDir())
	buf := &bytes.Buffer{}
	rt := &config.Runtime{}
	if err := Run(rt, buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := buf.String()
	for _, want := range []string{"drift:brew", "drift:dotfiles", "drift:macos"} {
		if !strings.Contains(got, want) {
			t.Errorf("default mode should show %q; got %q", want, got)
		}
	}
}

// TestRun_AlwaysReturnsNil verifies the read-only contract.
func TestRun_AlwaysReturnsNil(t *testing.T) {
	t.Setenv("MACHEIM_REPO", "")
	t.Setenv("HOME", t.TempDir())
	rt := &config.Runtime{}
	if err := Run(rt, &bytes.Buffer{}); err != nil {
		t.Errorf("status.Run must never return an error; got %v", err)
	}
}
