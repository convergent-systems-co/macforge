// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

//go:build integration

package verify_test

import (
	"context"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/apple"
	"github.com/convergent-systems-co/macforge/internal/apple/codesign"
	"github.com/convergent-systems-co/macforge/internal/apple/spctl"
	"github.com/convergent-systems-co/macforge/internal/verify"
)

func TestService_Verify_BothPass(t *testing.T) {
	r, _ := apple.NewFakeRunner("../../testdata")
	svc := verify.New(codesign.New(r), spctl.New(r))

	res, err := svc.Verify(context.Background(), "./build/MyApp.app", spctl.AssessTypeExec)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !res.Codesign.OK {
		t.Fatal("Codesign.OK = false, want true")
	}
	if !res.Spctl.Accepted {
		t.Fatal("Spctl.Accepted = false, want true")
	}
	if !res.OK {
		t.Fatal("Result.OK = false, want true")
	}
}
