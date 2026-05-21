// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package security

import (
	"reflect"
	"strings"
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

func TestArgs_ListKeychainsRead(t *testing.T) {
	got := argsListKeychainsRead()
	want := []string{"list-keychains", "-d", "user"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsListKeychainsRead = %v, want %v", got, want)
	}
}

func TestArgs_ListKeychainsWrite(t *testing.T) {
	paths := []string{
		"/Users/me/Library/Keychains/macforge-XYZ-signing.keychain-db",
		"/Users/me/Library/Keychains/login.keychain-db",
	}
	got := argsListKeychainsWrite(paths)
	want := []string{
		"list-keychains", "-d", "user", "-s",
		"/Users/me/Library/Keychains/macforge-XYZ-signing.keychain-db",
		"/Users/me/Library/Keychains/login.keychain-db",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argsListKeychainsWrite = %v, want %v", got, want)
	}
}

func TestParseSearchList(t *testing.T) {
	in := `    "/Users/itsfwcp/Library/Keychains/login.keychain-db"
    "/Library/Keychains/System.keychain"
`
	got := parseSearchList(in)
	want := []string{
		"/Users/itsfwcp/Library/Keychains/login.keychain-db",
		"/Library/Keychains/System.keychain",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSearchList = %v, want %v", got, want)
	}
}

func TestParseSearchList_IgnoresBlankLines(t *testing.T) {
	in := "\n    \"/a/b.keychain-db\"\n\n\n"
	got := parseSearchList(in)
	want := []string{"/a/b.keychain-db"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSearchList = %v, want %v", got, want)
	}
}

func TestKeychainPath_AddsSuffix(t *testing.T) {
	got, err := keychainPath("macforge-XYZ-signing")
	if err != nil {
		t.Fatalf("keychainPath: %v", err)
	}
	if !strings.HasSuffix(got, "/Library/Keychains/macforge-XYZ-signing.keychain-db") {
		t.Fatalf("got %q, want suffix /Library/Keychains/macforge-XYZ-signing.keychain-db", got)
	}
}

func TestKeychainPath_PreservesExplicitSuffix(t *testing.T) {
	got, err := keychainPath("macforge-XYZ-signing.keychain-db")
	if err != nil {
		t.Fatalf("keychainPath: %v", err)
	}
	if !strings.HasSuffix(got, "/Library/Keychains/macforge-XYZ-signing.keychain-db") {
		t.Fatalf("got %q, want path ending in single .keychain-db suffix", got)
	}
	if strings.Contains(got, ".keychain-db.keychain-db") {
		t.Fatalf("got %q: suffix was doubled", got)
	}
}
