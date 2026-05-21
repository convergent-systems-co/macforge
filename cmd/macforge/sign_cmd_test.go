// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "testing"

func TestSign_RequiresArtifact(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"apple", "sign"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing artifact arg")
	}
}
