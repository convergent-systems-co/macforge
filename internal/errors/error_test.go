// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package mferrors_test

import (
	stdErrors "errors"
	"strings"
	"testing"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func TestError_Format(t *testing.T) {
	err := mferrors.NewSigning(mferrors.CodeSignVerificationFail, "signing.Sign", "verification failed")
	if got, want := err.Error(), "[MF-SIGN-001] signing.Sign: verification failed"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestError_UnwrapPreservesCause(t *testing.T) {
	cause := stdErrors.New("original")
	wrapped := mferrors.NewSigning(mferrors.CodeSignVerificationFail, "signing.Sign", "failed",
		mferrors.WithCause(cause))
	if u := stdErrors.Unwrap(wrapped); u != cause {
		t.Fatalf("Unwrap = %v, want %v", u, cause)
	}
}

func TestError_IsSentinel(t *testing.T) {
	err := mferrors.NewSigning(mferrors.CodeSignVerificationFail, "signing.Sign", "msg")
	if !stdErrors.Is(err, mferrors.ErrSigning) {
		t.Fatal("errors.Is(err, ErrSigning) = false, want true")
	}
	if stdErrors.Is(err, mferrors.ErrKeychain) {
		t.Fatal("errors.Is(err, ErrKeychain) = true, want false (wrong subsystem)")
	}
}

func TestError_As(t *testing.T) {
	err := mferrors.NewNotarize(mferrors.CodeNotarizeRejected, "notarize.Submit", "rejected",
		mferrors.WithHint("Run macforge identity status"))

	var got *mferrors.Error
	if !stdErrors.As(err, &got) {
		t.Fatal("errors.As did not match *Error")
	}
	if got.Code != mferrors.CodeNotarizeRejected {
		t.Fatalf("Code = %q, want %q", got.Code, mferrors.CodeNotarizeRejected)
	}
	if got.Hint == "" || !strings.Contains(got.Hint, "identity status") {
		t.Fatalf("Hint = %q, want non-empty containing 'identity status'", got.Hint)
	}
}

func TestError_DetailsRedacted(t *testing.T) {
	err := mferrors.NewKeychain(mferrors.CodeKeychainLocked, "keychain.Unlock", "locked",
		mferrors.WithDetails(map[string]any{"keychain": "macforge-XYZ-signing"}))

	var got *mferrors.Error
	if !stdErrors.As(err, &got) {
		t.Fatal("errors.As failed")
	}
	if v, ok := got.Details["keychain"].(string); !ok || v != "macforge-XYZ-signing" {
		t.Fatalf("Details[keychain] = %v, want macforge-XYZ-signing", got.Details["keychain"])
	}
}

func TestCodes_AreUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, code := range mferrors.AllCodes() {
		if seen[code] {
			t.Fatalf("duplicate code: %s", code)
		}
		seen[code] = true
	}
}

func TestCodes_FollowFormat(t *testing.T) {
	for _, code := range mferrors.AllCodes() {
		if !strings.HasPrefix(code, "MF-") {
			t.Fatalf("code %q does not start with MF-", code)
		}
		parts := strings.Split(code, "-")
		if len(parts) != 3 {
			t.Fatalf("code %q does not match MF-<SUBSYSTEM>-NNN", code)
		}
	}
}
