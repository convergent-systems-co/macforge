// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package mferrors

// Constructors keep each error's subsystem sentinel correct by construction.
// Always prefer these over building *Error literals directly.

func newWithSentinel(sentinel error, code, op, msg string, opts ...Option) *Error {
	e := &Error{Code: code, Op: op, Msg: msg, Sentinel: sentinel}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// NewIdentity builds an *Error tagged with the identity subsystem sentinel.
// Use for cert/key/CSR lifecycle errors (codes prefixed MF-IDENT-).
func NewIdentity(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrIdentity, code, op, msg, opts...)
}

// NewKeychain builds an *Error tagged with the keychain subsystem sentinel.
// Use for keychain lifecycle errors (codes prefixed MF-KEYCHAIN-).
func NewKeychain(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrKeychain, code, op, msg, opts...)
}

// NewSigning builds an *Error tagged with the signing subsystem sentinel.
// Use for codesign-orchestration errors (codes prefixed MF-SIGN-).
func NewSigning(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrSigning, code, op, msg, opts...)
}

// NewPackage builds an *Error tagged with the packaging subsystem sentinel.
// Use for zip/dmg/pkg/app-bundle errors (codes prefixed MF-PACKAGE-).
func NewPackage(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrPackage, code, op, msg, opts...)
}

// NewNotarize builds an *Error tagged with the notarization subsystem sentinel.
// Use for notarytool / App Store Connect errors (codes prefixed MF-NOTARIZE-).
func NewNotarize(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrNotarize, code, op, msg, opts...)
}

// NewVerify builds an *Error tagged with the verification subsystem sentinel.
// Use for codesign --verify / spctl --assess errors (codes prefixed MF-VERIFY-).
func NewVerify(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrVerify, code, op, msg, opts...)
}

// NewPublish builds an *Error tagged with the publish subsystem sentinel.
// Use for GitHub Releases / artifact-upload errors (codes prefixed MF-PUBLISH-).
func NewPublish(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrPublish, code, op, msg, opts...)
}

// NewRelease builds an *Error tagged with the release-pipeline sentinel.
// Use for end-to-end orchestrator errors (codes prefixed MF-RELEASE-).
func NewRelease(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrRelease, code, op, msg, opts...)
}

// NewTool builds an *Error tagged with the apple-tool boundary sentinel.
// Use when an Apple CLI fails or is missing (codes prefixed MF-TOOL-).
func NewTool(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrTool, code, op, msg, opts...)
}

// NewConfig builds an *Error tagged with the config subsystem sentinel.
// Use for macforge.yaml parse / validation errors (codes prefixed MF-CONFIG-).
func NewConfig(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrConfig, code, op, msg, opts...)
}

// NewAudit builds an *Error tagged with the audit subsystem sentinel.
// Use for JSONL writer / redaction errors (codes prefixed MF-AUDIT-).
func NewAudit(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrAudit, code, op, msg, opts...)
}

// NewOutput builds an *Error tagged with the output subsystem sentinel.
// Use for renderer / envelope errors (codes prefixed MF-OUTPUT-).
func NewOutput(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrOutput, code, op, msg, opts...)
}

// NewCI builds an *Error tagged with the CI-detection subsystem sentinel.
// Use for CI provider integration errors (codes prefixed MF-CI-).
func NewCI(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrCI, code, op, msg, opts...)
}
