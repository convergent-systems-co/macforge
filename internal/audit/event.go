// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package audit

import "time"

// Kind enumerates the audit event kinds. Vocabulary deliberately mirrors
// ~/.ai/Common.md §5.2 for cross-tool grep consistency.
type Kind string

const (
	// KindInvocationAttempt is emitted just before an Apple tool is spawned.
	KindInvocationAttempt Kind = "invocation-attempt"
	// KindInvocationResult is emitted after the Apple tool returns; carries exit code + duration.
	KindInvocationResult Kind = "invocation-result"
	// KindSignal is a freeform notable event (start, end, milestone).
	KindSignal Kind = "signal"
	// KindDecision records a fork in MacForge's logic (e.g., chose identity X over Y).
	KindDecision Kind = "decision"
	// KindError records a structured failure (carries an MF-* code).
	KindError Kind = "error"
)

// Actor identifies who or what generated the event.
type Actor string

const (
	// ActorMacforge is MacForge itself (the default for runtime-generated events).
	ActorMacforge Actor = "macforge"
	// ActorUser is the human operator (events triggered by explicit invocation).
	ActorUser Actor = "user"
	// ActorTool is an external Apple/system CLI MacForge shelled out to.
	ActorTool Actor = "tool"
	// ActorSystem is the host OS or kernel.
	ActorSystem Actor = "system"
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
