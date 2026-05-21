// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package mferrors

import (
	stdErrors "errors"
	"fmt"
)

// Error is MacForge's structured error. It travels through three audiences:
// humans (via the renderer), CI scripts (via the JSON envelope), and the
// audit log (as a structured event).
type Error struct {
	Code     string         // MF-<SUBSYSTEM>-NNN
	Op       string         // caller site, e.g. "signing.Sign"
	Msg      string         // human-readable, terse
	Hint     string         // optional remediation suggestion
	Details  map[string]any // structured context (never secrets)
	Sentinel error          // subsystem sentinel; powers errors.Is matching
	Cause    error          // underlying cause; powers errors.Unwrap
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Op, e.Msg)
}

// Unwrap returns the underlying cause for errors.Unwrap and errors.Is.
func (e *Error) Unwrap() error {
	if e.Cause != nil {
		return e.Cause
	}
	return e.Sentinel
}

// Is implements errors.Is. It matches when the target is this error's
// subsystem sentinel, or when the target unwrap chain matches the cause.
func (e *Error) Is(target error) bool {
	if target == nil {
		return false
	}
	if e.Sentinel != nil && stdErrors.Is(e.Sentinel, target) {
		return true
	}
	if e.Cause != nil && stdErrors.Is(e.Cause, target) {
		return true
	}
	return false
}

// Option configures an Error during construction.
type Option func(*Error)

// WithCause attaches an underlying cause.
func WithCause(cause error) Option {
	return func(e *Error) { e.Cause = cause }
}

// WithHint attaches a remediation suggestion.
func WithHint(hint string) Option {
	return func(e *Error) { e.Hint = hint }
}

// WithDetails attaches structured context. Callers MUST NOT pass secret
// values; redaction happens at the Runner boundary, not here.
func WithDetails(details map[string]any) Option {
	return func(e *Error) {
		if e.Details == nil {
			e.Details = map[string]any{}
		}
		for k, v := range details {
			e.Details[k] = v
		}
	}
}
