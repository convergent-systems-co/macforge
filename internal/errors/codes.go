// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package mferrors

// Code values are part of the public MacForge contract. Once published,
// the meaning of a code never changes. Retired codes stay reserved.
// New conditions append; never re-use a number.
//
// Format: MF-<SUBSYSTEM>-NNN
//
// See docs/adr/0011-error-model-and-codes.md.
const (
	// Identity
	CodeIdentityNotFound   = "MF-IDENT-001"
	CodeIdentityExpired    = "MF-IDENT-002"
	CodeIdentityImportFail = "MF-IDENT-003"

	// Keychain
	CodeKeychainLocked          = "MF-KEYCHAIN-001"
	CodeKeychainMissing         = "MF-KEYCHAIN-002"
	CodeKeychainExists          = "MF-KEYCHAIN-003"
	CodeKeychainNonStandardName = "MF-KEYCHAIN-004"

	// Signing
	CodeSignVerificationFail        = "MF-SIGN-001"
	CodeSignNoIdentity              = "MF-SIGN-002"
	CodeSignHardenedRuntimeRequired = "MF-SIGN-003"

	// Packaging
	CodePackageUnsupportedFormat = "MF-PACKAGE-001"

	// Notarization
	CodeNotarizeRejected = "MF-NOTARIZE-001"
	CodeNotarizeTimeout  = "MF-NOTARIZE-002"

	// Verification
	CodeVerifyCodesignFail = "MF-VERIFY-001"
	CodeVerifySpctlFail    = "MF-VERIFY-002"

	// Publishing
	CodePublishUploadFail = "MF-PUBLISH-001"

	// Release pipeline
	CodeReleaseStepFail = "MF-RELEASE-001"

	// Tool boundary
	CodeToolMissing = "MF-TOOL-001"
	CodeToolFailed  = "MF-TOOL-002"

	// Config
	CodeConfigInvalid = "MF-CONFIG-001"
	CodeConfigMissing = "MF-CONFIG-002"

	// Audit
	CodeAuditWriteFail   = "MF-AUDIT-001"
	CodeAuditRedactPanic = "MF-AUDIT-002"

	// Output
	CodeOutputRenderFail = "MF-OUTPUT-001"

	// CI
	CodeCIProviderUnknown = "MF-CI-001"
)

// AllCodes returns every registered code. Tests use this to enforce
// uniqueness and format conformance.
func AllCodes() []string {
	return []string{
		CodeIdentityNotFound, CodeIdentityExpired, CodeIdentityImportFail,
		CodeKeychainLocked, CodeKeychainMissing, CodeKeychainExists, CodeKeychainNonStandardName,
		CodeSignVerificationFail, CodeSignNoIdentity, CodeSignHardenedRuntimeRequired,
		CodePackageUnsupportedFormat,
		CodeNotarizeRejected, CodeNotarizeTimeout,
		CodeVerifyCodesignFail, CodeVerifySpctlFail,
		CodePublishUploadFail,
		CodeReleaseStepFail,
		CodeToolMissing, CodeToolFailed,
		CodeConfigInvalid, CodeConfigMissing,
		CodeAuditWriteFail, CodeAuditRedactPanic,
		CodeOutputRenderFail,
		CodeCIProviderUnknown,
	}
}
