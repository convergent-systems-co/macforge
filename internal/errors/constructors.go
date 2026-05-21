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

func NewIdentity(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrIdentity, code, op, msg, opts...)
}

func NewKeychain(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrKeychain, code, op, msg, opts...)
}

func NewSigning(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrSigning, code, op, msg, opts...)
}

func NewPackage(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrPackage, code, op, msg, opts...)
}

func NewNotarize(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrNotarize, code, op, msg, opts...)
}

func NewVerify(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrVerify, code, op, msg, opts...)
}

func NewPublish(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrPublish, code, op, msg, opts...)
}

func NewRelease(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrRelease, code, op, msg, opts...)
}

func NewTool(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrTool, code, op, msg, opts...)
}

func NewConfig(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrConfig, code, op, msg, opts...)
}

func NewAudit(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrAudit, code, op, msg, opts...)
}

func NewOutput(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrOutput, code, op, msg, opts...)
}

func NewCI(code, op, msg string, opts ...Option) *Error {
	return newWithSentinel(ErrCI, code, op, msg, opts...)
}
