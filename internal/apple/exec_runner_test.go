// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package apple_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/audit"
)

func TestExecRunner_CapturesStdout(t *testing.T) {
	r := apple.NewExecRunner(nil)
	res, err := r.Run(context.Background(), apple.Invocation{
		Tool: "echo",
		Args: []string{"hello-macforge"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", res.ExitCode)
	}
	if !strings.Contains(string(res.Stdout), "hello-macforge") {
		t.Fatalf("Stdout = %q, want to contain 'hello-macforge'", string(res.Stdout))
	}
}

func TestExecRunner_ToolMissing(t *testing.T) {
	r := apple.NewExecRunner(nil)
	_, err := r.Run(context.Background(), apple.Invocation{
		Tool: "definitely-not-a-real-binary-xyz123",
	})
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
}

func TestExecRunner_Timeout(t *testing.T) {
	r := apple.NewExecRunner(nil)
	_, err := r.Run(context.Background(), apple.Invocation{
		Tool:    "sleep",
		Args:    []string{"5"},
		Timeout: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestExecRunner_RedactsInvocationSecrets is the regression test for issue #3.
// Per Common.md §4 and ADR-0012, any secret declared in Invocation.Redact MUST
// NOT appear in the audit log's probe_payload.
func TestExecRunner_RedactsInvocationSecrets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.jsonl")
	w, err := audit.NewWriter(path, audit.NewRedactor(nil)) // writer has NO secrets known
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	r := apple.NewExecRunner(w)
	r.SetTrace("TEST-REDACT")

	const password = "hunter2-super-secret-xyz"
	_, err = r.Run(context.Background(), apple.Invocation{
		Tool:   "echo",
		Args:   []string{"--password", password, "rest"},
		Redact: []string{password},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if strings.Contains(string(contents), password) {
		t.Fatalf("audit log leaked secret %q:\n%s", password, string(contents))
	}
	if !strings.Contains(string(contents), "[REDACTED]") {
		t.Fatalf("audit log missing [REDACTED] marker:\n%s", string(contents))
	}
}
