// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package apple_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/convergent-systems-co/macforge/internal/apple"
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
