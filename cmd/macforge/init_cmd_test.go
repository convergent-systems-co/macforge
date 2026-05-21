// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_WritesGlobalScaffold(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	root := newRootCmd()
	root.SetArgs([]string{"init", "--team", "XYZ1234567"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	path := filepath.Join(dir, "macforge", "macforge.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected macforge.yaml at %s: %v", path, err)
	}
	content := string(b)
	for _, want := range []string{
		"version: 1",
		"team: XYZ1234567",
		"keychain:",
		"macforge-XYZ1234567-signing",
		"identity:",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("scaffold missing %q\n---\n%s", want, content)
		}
	}

	// Scaffold should NOT contain artifact-shaped fields (per ADR-0015):
	for _, badField := range []string{
		"entitlements:",
		"package:",
		"publish:",
		"github:",
	} {
		if strings.Contains(content, badField) {
			t.Errorf("scaffold contains project-shaped field %q; those belong in ./macforge.yaml\n---\n%s", badField, content)
		}
	}
}

func TestInit_RefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Pre-seed the global config.
	cfgDir := filepath.Join(dir, "macforge")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "macforge.yaml"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	root := newRootCmd()
	root.SetArgs([]string{"init", "--team", "XYZ"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected refusal when global macforge.yaml exists")
	}
}

func TestInit_CreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	// Parent dir doesn't exist yet — init must MkdirAll it.

	root := newRootCmd()
	root.SetArgs([]string{"init", "--team", "ABC9876543"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	cfgDir := filepath.Join(dir, "macforge")
	st, err := os.Stat(cfgDir)
	if err != nil {
		t.Fatalf("parent dir not created: %v", err)
	}
	if !st.IsDir() {
		t.Fatalf("%s is not a directory", cfgDir)
	}
}
