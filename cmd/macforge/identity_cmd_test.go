// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestIdentity_HelpListsSubverbs(t *testing.T) {
	var buf bytes.Buffer
	root := newRootCmd()
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"identity", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	for _, sub := range []string{"import", "list", "status"} {
		if !strings.Contains(buf.String(), sub) {
			t.Errorf("missing subverb %q\n%s", sub, buf.String())
		}
	}
}
