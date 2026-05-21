// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"strings"
	"testing"
)

func TestWorkstation_DoctorRunsAndExits(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"workstation", "doctor"})
	// doctor may exit non-zero on this host (no Homebrew, no repo, etc.).
	// We just verify it runs without panicking and either succeeds or
	// returns a non-cobra-internal error.
	if err := root.Execute(); err != nil {
		// Acceptable: doctor checks failed (exit 1 path). Verify the error
		// is the sentinel, not something like "unknown command".
		if !strings.Contains(err.Error(), "doctor") {
			t.Fatalf("doctor returned unexpected error: %v", err)
		}
	}
}

func TestWorkstation_StubsPrintNotImplemented(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"workstation", "bootstrap"}, "not implemented yet"},
		{[]string{"workstation", "zsh", "setup"}, "not implemented yet"},
		{[]string{"workstation", "macos", "defaults"}, "not implemented yet"},
		{[]string{"workstation", "downloads"}, "not implemented yet"},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			root := newRootCmd()
			var sb strings.Builder
			root.SetOut(&sb)
			root.SetErr(&sb)
			root.SetArgs(tc.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if !strings.Contains(sb.String(), tc.want) {
				t.Errorf("output missing %q\n---\n%s", tc.want, sb.String())
			}
		})
	}
}

func TestWorkstation_StatusRendersWithoutPanic(t *testing.T) {
	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetErr(&sb)
	root.SetArgs([]string{"workstation", "status"})
	// status is read-only and tolerates an unconfigured environment;
	// it should never return a non-nil error.
	if err := root.Execute(); err != nil {
		t.Fatalf("status returned unexpected error: %v", err)
	}
}
