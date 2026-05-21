// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRoot_HelpListsAppleAndVersion(t *testing.T) {
	// Per ADR-0017, the root surface is just `apple` (today's verbs) and
	// `version`. All Apple-platform verbs are nested under `apple`.
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	for _, top := range []string{"apple", "version"} {
		if !strings.Contains(out, top) {
			t.Errorf("--help output missing top-level command %q\n%s", top, out)
		}
	}
}

func TestApple_HelpListsAllVerbs(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"apple", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	out := buf.String()
	for _, verb := range []string{"init", "identity", "keychain", "sign", "package", "notarize", "verify", "publish", "release", "audit"} {
		if !strings.Contains(out, verb) {
			t.Errorf("apple --help missing verb %q\n%s", verb, out)
		}
	}
}

func TestRoot_GlobalFlagsPresent(t *testing.T) {
	root := newRootCmd()
	for _, flag := range []string{"config", "output", "log-level", "team-id", "dry-run", "no-color", "verbose"} {
		if root.PersistentFlags().Lookup(flag) == nil {
			t.Errorf("global flag --%s not registered", flag)
		}
	}
}

func TestRoot_VersionCommand(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(buf.String(), "macforge") {
		t.Fatalf("version output missing 'macforge': %q", buf.String())
	}
}
