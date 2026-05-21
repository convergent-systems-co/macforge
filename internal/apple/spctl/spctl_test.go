// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package spctl

import (
	"reflect"
	"testing"
)

func TestArgs_AssessExec(t *testing.T) {
	got := argsAssess("./MyApp.app", AssessTypeExec)
	want := []string{"--assess", "--type", "execute", "--verbose=4", "./MyApp.app"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsAssess = %v, want %v", got, want)
	}
}
