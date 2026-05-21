// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build integration

package spctl_test

import (
	"context"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/spctl"
)

func TestAssess_Accept(t *testing.T) {
	r, _ := apple.NewFakeRunner("../../../testdata")
	c := spctl.New(r)

	res, err := c.Assess(context.Background(), "./build/MyApp.app", spctl.AssessTypeExec)
	if err != nil {
		t.Fatalf("Assess: %v", err)
	}
	if !res.Accepted {
		t.Fatal("Accepted = false, want true")
	}
	if res.Source == "" {
		t.Fatalf("Source = %q, want non-empty", res.Source)
	}
}
