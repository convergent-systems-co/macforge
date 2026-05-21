package cmd

import (
	"strings"
	"testing"
)

func TestUpdateLocalToRemote_UnknownModule(t *testing.T) {
	t.Parallel()
	// In a clean working dir there is no repo configured, so brew /
	// dotfiles handlers would refuse with the read-only error. We test
	// the dispatcher's module validation directly via an invalid value
	// — that path returns before any module runs.
	_, _, _, err := runRoot(t, "update", "local-to-remote", "--module", "bogus")
	if err == nil {
		t.Fatal("expected error for unknown module, got nil")
	}
	if !strings.Contains(err.Error(), "unknown module") {
		t.Errorf("error %q missing 'unknown module'", err.Error())
	}
}

func TestUpdateRemoteToLocal_UnknownModule(t *testing.T) {
	t.Parallel()
	_, _, _, err := runRoot(t, "update", "remote-to-local", "--module", "bogus")
	if err == nil {
		t.Fatal("expected error for unknown module, got nil")
	}
	if !strings.Contains(err.Error(), "unknown module") {
		t.Errorf("error %q missing 'unknown module'", err.Error())
	}
}

func TestUpdateLocalToRemote_DefaultsModuleStubsCleanly(t *testing.T) {
	t.Parallel()
	// Scope to defaults so neither brew nor dotfiles is exercised. The
	// stub prints to stderr and returns nil — no error to the dispatcher.
	// The defaults stub doesn't require a repo to be configured.
	_, _, _, err := runRoot(t, "update", "local-to-remote", "--module", "defaults")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateRemoteToLocal_DefaultsModuleStubsCleanly(t *testing.T) {
	t.Parallel()
	_, _, _, err := runRoot(t, "update", "remote-to-local", "--module", "defaults")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDispatchModules_All(t *testing.T) {
	t.Parallel()
	got := dispatchModules("all")
	want := []string{"brew", "dotfiles", "defaults"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dispatchModules(all)[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestDispatchModules_Single(t *testing.T) {
	t.Parallel()
	for _, m := range []string{"brew", "dotfiles", "defaults"} {
		got := dispatchModules(m)
		if len(got) != 1 || got[0] != m {
			t.Errorf("dispatchModules(%q) = %v, want [%q]", m, got, m)
		}
	}
}
