// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package apple

import "time"

// Invocation describes one Apple-tool call. It is the only shape that crosses
// the Runner boundary. Args is argv (no shell interpolation). Redact lists
// substrings to mask in the audit log.
type Invocation struct {
	Tool    string
	Args    []string
	Stdin   []byte
	Env     map[string]string
	Timeout time.Duration
	Redact  []string
}

// Result is what a Runner returns from a successful spawn. ExitCode and
// the captured streams are always populated; a non-zero ExitCode is not
// a Go error — callers decide whether non-zero is an error in context.
type Result struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	Duration time.Duration
}
