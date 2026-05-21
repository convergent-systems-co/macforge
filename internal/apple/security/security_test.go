// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package security

import (
	"reflect"
	"testing"
)

func TestArgs_CreateKeychain(t *testing.T) {
	got := argsCreateKeychain("macforge-XYZ-signing", "supersecret")
	want := []string{"create-keychain", "-p", "supersecret", "macforge-XYZ-signing"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsCreateKeychain = %v, want %v", got, want)
	}
}

func TestArgs_UnlockKeychain(t *testing.T) {
	got := argsUnlockKeychain("macforge-XYZ-signing", "supersecret")
	want := []string{"unlock-keychain", "-p", "supersecret", "macforge-XYZ-signing"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsUnlockKeychain = %v, want %v", got, want)
	}
}

func TestArgs_SetKeychainSettings_LockAndTimeout(t *testing.T) {
	got := argsSetKeychainSettings("macforge-XYZ-signing", true, 3600)
	want := []string{"set-keychain-settings", "-l", "-t", "3600", "macforge-XYZ-signing"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsSetKeychainSettings = %v, want %v", got, want)
	}
}

func TestArgs_FindIdentity(t *testing.T) {
	got := argsFindIdentity("macforge-XYZ-signing", "codesigning")
	want := []string{"find-identity", "-p", "codesigning", "-v", "macforge-XYZ-signing"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsFindIdentity = %v, want %v", got, want)
	}
}
