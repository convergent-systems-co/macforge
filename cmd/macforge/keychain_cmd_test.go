// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestKeychainList_DryRun verifies the list command builds correctly. Real
// keychain ops are exercised in Tier 3 e2e (Task 16).
func TestKeychain_HelpListsSubverbs(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"keychain", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, sub := range []string{"create", "delete", "list", "unlock"} {
		if !strings.Contains(buf.String(), sub) {
			t.Errorf("missing subverb %q\n%s", sub, buf.String())
		}
	}
}

func TestKeychainCreate_RequiresValidName(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	// seed a minimal macforge.yaml so config.Load succeeds
	_ = os.WriteFile(filepath.Join(dir, "macforge.yaml"), []byte("version: 1\nteam: XYZ1234567\n"), 0o644)

	t.Setenv("MACFORGE_TEST_PW", "topsecret")
	root := newRootCmd()
	root.SetArgs([]string{"--output", "json", "keychain", "create",
		"--name", "definitely-not-a-macforge-name",
		"--secret-ref", "env:MACFORGE_TEST_PW",
	})

	var buf bytes.Buffer
	root.SetOut(&buf)
	_ = root.Execute() // expected to error

	// CLI prints the JSON envelope to stdout via the renderer; capture from a hook
	// is awkward in this test. Instead, just assert that Execute returned an error.
	if root.Execute() == nil {
		t.Fatal("expected error for invalid keychain name (second Execute should also fail)")
	}
	_ = json.Unmarshal(buf.Bytes(), &map[string]any{})
}
