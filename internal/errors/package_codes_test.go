// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package mferrors_test

// Tests for the MF-PACKAGE-* error codes introduced in T1 (#21).
//
// These tests are intentionally written against raw string literals rather
// than the named constants (e.g. "MF-PACKAGE-001" instead of
// mferrors.CodePackageInputNotFound). The constants don't exist yet; the
// tests fail until the T1 coder adds them and registers them in AllCodes().

import (
	stdErrors "errors"
	"testing"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// TestNewPackage_SentinelEachNewCode verifies that mferrors.NewPackage built
// with any of the five new MF-PACKAGE-* codes satisfies errors.Is(err,
// mferrors.ErrPackage) and that each code is registered in AllCodes().
//
// Each sub-test will fail until T1's coder adds the constants and populates
// AllCodes() with them.
func TestNewPackage_SentinelEachNewCode(t *testing.T) {
	t.Parallel()

	codes := []struct {
		code    string
		comment string
	}{
		{"MF-PACKAGE-001", "CodePackageInputNotFound — input path missing"},
		{"MF-PACKAGE-002", "CodePackageInputNotAppBundle — input is not a .app bundle"},
		{"MF-PACKAGE-003", "CodePackageInputNotSigned — input fails codesign --verify"},
		{"MF-PACKAGE-004", "CodePackageOutputExists — output path occupied and --force not set"},
		{"MF-PACKAGE-005", "CodePackageDittoFailed — ditto subprocess exited non-zero"},
	}

	for _, tc := range codes {
		tc := tc // capture loop variable
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()

			err := mferrors.NewPackage(tc.code, "test.op", "test message")

			// 1. The returned error must satisfy errors.Is(err, ErrPackage).
			if !stdErrors.Is(err, mferrors.ErrPackage) {
				t.Errorf("errors.Is(NewPackage(%q, ...), ErrPackage) = false; want true (%s)", tc.code, tc.comment)
			}

			// 2. The code must NOT satisfy errors.Is for a different subsystem
			// sentinel (guard against accidentally tagging the wrong sentinel).
			if stdErrors.Is(err, mferrors.ErrSigning) {
				t.Errorf("errors.Is(NewPackage(%q, ...), ErrSigning) = true; want false — wrong subsystem sentinel", tc.code)
			}

			// 3. The code must appear in AllCodes() — registration check.
			found := false
			for _, c := range mferrors.AllCodes() {
				if c == tc.code {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("code %q not present in AllCodes(); add it to the AllCodes() return slice in codes.go", tc.code)
			}
		})
	}
}

// TestCode_001_IsInputNotFound pins that MF-PACKAGE-001 means
// CodePackageInputNotFound — NOT CodePackageUnsupportedFormat (which was
// removed in T1 because it was unused dead code). This is a regression guard:
// if someone accidentally resurrects the old constant, this test fails.
func TestCode_001_IsInputNotFound(t *testing.T) {
	t.Parallel()

	if mferrors.CodePackageInputNotFound != "MF-PACKAGE-001" {
		t.Errorf("CodePackageInputNotFound = %q, want %q — slot 001 must be InputNotFound, not UnsupportedFormat",
			mferrors.CodePackageInputNotFound, "MF-PACKAGE-001")
	}
}

// TestNewPackage_As verifies that errors.As correctly extracts the underlying
// *mferrors.Error from a NewPackage-built error, and that Code and Op are
// populated faithfully for each new code.
func TestNewPackage_As(t *testing.T) {
	t.Parallel()

	cases := []struct {
		code string
		op   string
		msg  string
	}{
		{"MF-PACKAGE-001", "packaging.Package", "no such file or directory: /tmp/missing.app"},
		{"MF-PACKAGE-002", "packaging.Package", "input is not a .app bundle"},
		{"MF-PACKAGE-003", "packaging.Package", "codesign --verify failed"},
		{"MF-PACKAGE-004", "packaging.Package", "output path already exists"},
		{"MF-PACKAGE-005", "ditto.Archive", "ditto exited with code 1"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()

			err := mferrors.NewPackage(tc.code, tc.op, tc.msg)

			var got *mferrors.Error
			if !stdErrors.As(err, &got) {
				t.Fatalf("errors.As(*mferrors.Error) = false for code %s", tc.code)
			}
			if got.Code != tc.code {
				t.Errorf("Error.Code = %q, want %q", got.Code, tc.code)
			}
			if got.Op != tc.op {
				t.Errorf("Error.Op = %q, want %q", got.Op, tc.op)
			}
			if got.Msg != tc.msg {
				t.Errorf("Error.Msg = %q, want %q", got.Msg, tc.msg)
			}
		})
	}
}
