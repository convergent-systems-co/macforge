// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package apple_test

import (
	"context"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
)

func TestFakeRunner_ReplaysFixture(t *testing.T) {
	r, err := apple.NewFakeRunner("../../testdata")
	if err != nil {
		t.Fatalf("NewFakeRunner: %v", err)
	}

	res, err := r.Run(context.Background(), apple.Invocation{
		Tool: "echo",
		Args: []string{"hello"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("ExitCode = %d, want 0", res.ExitCode)
	}
	if string(res.Stdout) != "hello\n" {
		t.Fatalf("Stdout = %q, want %q", string(res.Stdout), "hello\n")
	}
}

func TestFakeRunner_UnmatchedInvocationFails(t *testing.T) {
	r, err := apple.NewFakeRunner("../../testdata")
	if err != nil {
		t.Fatalf("NewFakeRunner: %v", err)
	}
	_, err = r.Run(context.Background(), apple.Invocation{Tool: "codesign", Args: []string{"--never-recorded"}})
	if err == nil {
		t.Fatal("expected error for unmatched fixture, got nil")
	}
}
