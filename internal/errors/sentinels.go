// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package mferrors

import stdErrors "errors"

// Sentinel roots — one per subsystem. Use errors.Is(err, ErrSigning) to match
// at coarse granularity. For code-level matching, use errors.As to extract
// *Error and compare its Code field.
var (
	ErrIdentity = stdErrors.New("identity")
	ErrKeychain = stdErrors.New("keychain")
	ErrSigning  = stdErrors.New("signing")
	ErrPackage  = stdErrors.New("package")
	ErrNotarize = stdErrors.New("notarize")
	ErrVerify   = stdErrors.New("verify")
	ErrPublish  = stdErrors.New("publish")
	ErrRelease  = stdErrors.New("release")
	ErrTool     = stdErrors.New("tool")
	ErrConfig   = stdErrors.New("config")
	ErrAudit    = stdErrors.New("audit")
	ErrOutput   = stdErrors.New("output")
	ErrCI       = stdErrors.New("ci")
)
