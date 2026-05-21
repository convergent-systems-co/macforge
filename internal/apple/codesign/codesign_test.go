// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package codesign

import (
	"reflect"
	"testing"
)

func TestArgs_SignFull(t *testing.T) {
	got := argsSign(SignOptions{
		IdentityCN:      "Developer ID Application: ACME (XYZ1234567)",
		Keychain:        "macforge-XYZ1234567-signing",
		HardenedRuntime: true,
		Timestamp:       true,
		Entitlements:    "./Entitlements.plist",
		Path:            "./build/MyApp.app",
	})
	want := []string{
		"--sign", "Developer ID Application: ACME (XYZ1234567)",
		"--keychain", "macforge-XYZ1234567-signing",
		"--options", "runtime",
		"--timestamp",
		"--entitlements", "./Entitlements.plist",
		"--verbose",
		"./build/MyApp.app",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsSign = %v\nwant %v", got, want)
	}
}

func TestArgs_Verify(t *testing.T) {
	got := argsVerify("./build/MyApp.app")
	want := []string{"--verify", "--strict", "--verbose=2", "./build/MyApp.app"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsVerify = %v, want %v", got, want)
	}
}

func TestArgs_Display(t *testing.T) {
	got := argsDisplay("./build/MyApp.app")
	want := []string{"--display", "--verbose=4", "./build/MyApp.app"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsDisplay = %v, want %v", got, want)
	}
}
