// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package audit

import "time"

// Kind enumerates the audit event kinds. Vocabulary deliberately mirrors
// ~/.ai/Common.md §5.2 for cross-tool grep consistency.
type Kind string

const (
	KindInvocationAttempt Kind = "invocation-attempt"
	KindInvocationResult  Kind = "invocation-result"
	KindSignal            Kind = "signal"
	KindDecision          Kind = "decision"
	KindError             Kind = "error"
)

// Actor identifies who or what generated the event.
type Actor string

const (
	ActorMacforge Actor = "macforge"
	ActorUser     Actor = "user"
	ActorTool     Actor = "tool"
	ActorSystem   Actor = "system"
)

// Event is a single audit-log entry. Marshals to one JSONL line.
// Field order in JSON is determined by struct order via encoding/json.
type Event struct {
	Chronon        time.Time      `json:"chronon"`
	Trace          string         `json:"trace"`
	Cwd            string         `json:"cwd"`
	Actor          Actor          `json:"actor"`
	Kind           Kind           `json:"kind"`
	Probe          string         `json:"probe,omitempty"`
	ProbePayload   string         `json:"probe_payload,omitempty"`
	Exit           *int           `json:"exit,omitempty"`
	DurationMs     int64          `json:"duration_ms,omitempty"`
	StdoutSHA256   string         `json:"stdout_sha256,omitempty"`
	StderrSHA256   string         `json:"stderr_sha256,omitempty"`
	ArtifactSHA256 string         `json:"artifact_sha256,omitempty"`
	Redacted       []string       `json:"redacted,omitempty"`
	Note           string         `json:"note,omitempty"`
	Code           string         `json:"code,omitempty"`
	Op             string         `json:"op,omitempty"`
	Team           string         `json:"team,omitempty"`
	IdentitySHA1   string         `json:"identity_sha1,omitempty"`
	Extra          map[string]any `json:"-"` // not emitted; for typed helpers
}
