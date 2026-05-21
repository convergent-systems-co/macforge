// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package audit

import "strings"

// Redactor replaces declared secret substrings with [REDACTED] before any
// audit line is written. Per ~/.ai/Common.md §4, secret values must never
// appear in audit logs. The Runner declares the secrets it knows about
// in Invocation.Redact; this type applies them.
type Redactor struct {
	secrets []string
}

// NewRedactor returns a Redactor that masks each non-empty entry of secrets.
func NewRedactor(secrets []string) *Redactor {
	clean := make([]string, 0, len(secrets))
	for _, s := range secrets {
		if s != "" {
			clean = append(clean, s)
		}
	}
	return &Redactor{secrets: clean}
}

// Apply returns input with every known secret substring replaced by [REDACTED].
func (r *Redactor) Apply(input string) string {
	out := input
	for _, s := range r.secrets {
		out = strings.ReplaceAll(out, s, "[REDACTED]")
	}
	return out
}
