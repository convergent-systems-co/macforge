// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package output

// successEnvelope is the wire shape for --output json on success.
// Field order matters for downstream consumers; keep stable.
type successEnvelope struct {
	OK      bool   `json:"ok"`
	Schema  string `json:"schema"`
	Trace   string `json:"trace"`
	Command string `json:"command"`
	Result  any    `json:"result"`
}

// errorEnvelope is the wire shape for --output json on failure.
type errorEnvelope struct {
	OK      bool           `json:"ok"`
	Schema  string         `json:"schema"`
	Trace   string         `json:"trace"`
	Command string         `json:"command"`
	Code    string         `json:"code"`
	Op      string         `json:"op"`
	Message string         `json:"message"`
	Hint    string         `json:"hint,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}
