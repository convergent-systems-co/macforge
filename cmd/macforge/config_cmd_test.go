// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfig_HelpListsValidate verifies the `apple config` subtree is
// registered under `apple` and lists the `validate` verb.
func TestConfig_HelpListsValidate(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"apple", "config", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(buf.String(), "validate") {
		t.Errorf("apple config --help missing 'validate':\n%s", buf.String())
	}
}

// TestConfigValidate_FailsOnMissingTeam confirms the verb propagates the
// strict static validator's verdict: a config with no team makes
// config.Load return an error, which the verb surfaces via rt.emit as a
// non-zero-exit failure envelope.
func TestConfigValidate_FailsOnMissingTeam(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	// Seed a global config that violates the new "team is required" rule.
	cfgDir := filepath.Join(dir, "macforge")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "macforge.yaml"),
		[]byte("version: 1\n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Project-local override must NOT exist in the test's cwd, or it could
	// inject a team via merge. t.Chdir keeps us out of the dev cwd.
	t.Chdir(t.TempDir())

	root := newRootCmd()
	root.SetArgs([]string{"apple", "config", "validate"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit when team is missing, got nil")
	}
	if !strings.Contains(err.Error(), "team") {
		t.Fatalf("error should mention 'team' field; got: %v", err)
	}
}

// TestConfigValidate_FailsOnTeamMismatch reproduces the #13 bug at the
// verb level: a stale team segment in keychain.name produces a Load
// failure (MF-CONFIG-001) with a message naming both teams, which the
// verb surfaces via rt.emit.
//
// The assertion checks the error message contains BOTH team strings;
// that is selective for the team-mismatch validator and would not be
// satisfied by an incidental Load failure (missing keychain file, etc.).
func TestConfigValidate_FailsOnTeamMismatch(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfgDir := filepath.Join(dir, "macforge")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	yaml := `version: 1
team: XYZ1234567
keychain:
  name: macforge-ABC9876543-signing
`
	if err := os.WriteFile(filepath.Join(cfgDir, "macforge.yaml"),
		[]byte(yaml), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	t.Chdir(t.TempDir())

	root := newRootCmd()
	root.SetArgs([]string{"apple", "config", "validate"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected non-zero exit when keychain.name disagrees with team, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "XYZ1234567") || !strings.Contains(msg, "ABC9876543") {
		t.Fatalf("error should name both conflicting team values; got: %v", msg)
	}
}

// TestConfigValidate_FailsWhenUnlockEnvUnset proves the runtime env-var
// check fires: a valid config that references env:MACFORGE_TEST_UNSET_VAR
// produces a non-zero exit AND the rendered output mentions the unset
// variable by name. The latter assertion is selective for the env-var
// runtime check (a missing keychain or absent security binary would also
// produce non-zero exit but would not emit "MACFORGE_TEST_UNSET_VAR").
func TestConfigValidate_FailsWhenUnlockEnvUnset(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	cfgDir := filepath.Join(dir, "macforge")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.Unsetenv("MACFORGE_TEST_UNSET_VAR"); err != nil {
		t.Fatalf("Unsetenv: %v", err)
	}
	yaml := `version: 1
team: XYZ1234567
keychain:
  name: macforge-XYZ1234567-signing
  unlock: env:MACFORGE_TEST_UNSET_VAR
`
	if err := os.WriteFile(filepath.Join(cfgDir, "macforge.yaml"),
		[]byte(yaml), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	t.Chdir(t.TempDir())

	// Swap stdoutForRenderer so we can read the verb's output.
	var stdoutBuf bytes.Buffer
	origStdout := stdoutForRenderer
	stdoutForRenderer = &stdoutBuf
	t.Cleanup(func() { stdoutForRenderer = origStdout })

	// Force human output mode regardless of CI default — the env-var line
	// goes into HumanLines, not into the failure envelope. (Under
	// GITHUB_ACTIONS the default is json, which would mask the assertion.)
	root := newRootCmd()
	root.SetArgs([]string{"--output", "human", "apple", "config", "validate"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected non-zero exit when keychain.unlock env var is unset, got nil")
	}
	output := stdoutBuf.String()
	if !strings.Contains(output, "MACFORGE_TEST_UNSET_VAR") {
		t.Fatalf("verb output should mention the unset env var by name; got:\n%s", output)
	}
}

// TestConfigValidateResult_SchemaName locks the JSON schema string.
// Per ADR-0017 (and ADR-0019 / this PR) every apple-subtree result uses
// the macforge.v1.apple.<verb> shape; an accidental rename to
// macforge.v1.config.validate would silently break downstream JSON
// consumers.
func TestConfigValidateResult_SchemaName(t *testing.T) {
	r := configValidateResult{}
	if got, want := r.SchemaName(), "macforge.v1.apple.config.validate"; got != want {
		t.Fatalf("SchemaName = %q, want %q", got, want)
	}
}

// TestConfigValidateResult_HumanLines renders the three marker forms
// (ok, fail, info) and verifies the prefixes the operator will see.
func TestConfigValidateResult_HumanLines(t *testing.T) {
	r := configValidateResult{
		Checks: []validateCheck{
			checkOK("alpha"),
			checkFail("beta", "do the thing"),
			checkInfo("gamma"),
		},
		Errors: 1,
	}
	lines := r.HumanLines()
	joined := strings.Join(lines, "\n")
	for _, want := range []string{
		"✓ alpha",
		"✗ beta",
		"hint: do the thing",
		"○ gamma",
		"1 errors",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("HumanLines missing %q:\n%s", want, joined)
		}
	}
}
