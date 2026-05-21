// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_WritesScaffold(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	root := newRootCmd()
	root.SetArgs([]string{"init", "--team", "XYZ1234567"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	path := filepath.Join(dir, "macforge.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected macforge.yaml at %s: %v", path, err)
	}
	content := string(b)
	for _, want := range []string{"version: 1", "team: XYZ1234567", "keychain:", "macforge-XYZ1234567-signing"} {
		if !strings.Contains(content, want) {
			t.Errorf("scaffold missing %q\n%s", want, content)
		}
	}
}

func TestInit_RefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "macforge.yaml"), []byte("existing"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	root := newRootCmd()
	root.SetArgs([]string{"init", "--team", "XYZ"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected refusal when macforge.yaml exists")
	}
}
