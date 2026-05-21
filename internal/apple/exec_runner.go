// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package apple

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"

	"github.com/convergent-systems-co/macforge/internal/audit"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// ExecRunner is the production Runner. It spawns real processes via os/exec,
// writes audit events on attempt and result, and applies the Invocation's
// Redact list via the audit Writer.
type ExecRunner struct {
	audit *audit.Writer // may be nil (no-op auditing)
	trace string        // trace ID stamped into every audit event
}

// NewExecRunner returns an ExecRunner that writes audit events through w.
// Pass nil to disable auditing (useful in tests).
func NewExecRunner(w *audit.Writer) *ExecRunner {
	return &ExecRunner{audit: w}
}

// SetTrace stamps trace onto every audit event this runner emits.
func (r *ExecRunner) SetTrace(trace string) { r.trace = trace }

// Run executes inv. Returns Result on successful spawn (regardless of exit
// code); returns *mferrors.Error for spawn failures (tool missing, context
// cancelled, timeout).
func (r *ExecRunner) Run(ctx context.Context, inv Invocation) (Result, error) {
	if inv.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, inv.Timeout)
		defer cancel()
	}

	r.writeAttempt(inv)

	cmd := exec.CommandContext(ctx, inv.Tool, inv.Args...)
	for k, v := range inv.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if len(inv.Stdin) > 0 {
		cmd.Stdin = bytes.NewReader(inv.Stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	dur := time.Since(start)

	res := Result{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		Duration: dur,
	}

	if runErr != nil {
		// If the context expired or was cancelled, treat it as a spawn failure
		// (not a non-zero exit code from the process itself).
		if ctx.Err() != nil {
			r.writeResult(inv, Result{Duration: dur})
			return Result{}, mferrors.NewTool(mferrors.CodeToolMissing, "apple.ExecRunner.Run",
				"failed to invoke "+inv.Tool, mferrors.WithCause(runErr))
		}
		var execErr *exec.ExitError
		if errors.As(runErr, &execErr) {
			res.ExitCode = execErr.ExitCode()
			r.writeResult(inv, res)
			return res, nil
		}
		r.writeResult(inv, Result{Duration: dur})
		return Result{}, mferrors.NewTool(mferrors.CodeToolMissing, "apple.ExecRunner.Run",
			"failed to invoke "+inv.Tool, mferrors.WithCause(runErr))
	}

	res.ExitCode = 0
	r.writeResult(inv, res)
	return res, nil
}

func (r *ExecRunner) writeAttempt(inv Invocation) {
	if r.audit == nil {
		return
	}
	cwd, _ := osGetwd()
	payload := argvString(inv)
	// Apply the per-invocation Redact list BEFORE the event reaches the Writer.
	// The Writer has its own redactor for cross-cutting secrets declared at
	// construction time; this handles secrets known only to a single call site
	// (e.g. keychain passwords, --p12-password values). See issue #3.
	if len(inv.Redact) > 0 {
		payload = audit.NewRedactor(inv.Redact).Apply(payload)
	}
	_ = r.audit.Write(audit.Event{
		Trace:        r.trace,
		Cwd:          cwd,
		Actor:        audit.ActorMacforge,
		Kind:         audit.KindInvocationAttempt,
		Probe:        inv.Tool,
		ProbePayload: payload,
		Redacted:     redactedKinds(inv),
	})
}

func (r *ExecRunner) writeResult(inv Invocation, res Result) {
	if r.audit == nil {
		return
	}
	cwd, _ := osGetwd()
	exit := res.ExitCode
	_ = r.audit.Write(audit.Event{
		Trace:        r.trace,
		Cwd:          cwd,
		Actor:        audit.ActorMacforge,
		Kind:         audit.KindInvocationResult,
		Probe:        inv.Tool,
		Exit:         &exit,
		DurationMs:   res.Duration.Milliseconds(),
		StdoutSHA256: sha256Hex(res.Stdout),
		StderrSHA256: sha256Hex(res.Stderr),
		Redacted:     redactedKinds(inv),
	})
}

func argvString(inv Invocation) string {
	out := inv.Tool
	for _, a := range inv.Args {
		out += " " + a
	}
	return out
}

func redactedKinds(inv Invocation) []string {
	if len(inv.Redact) == 0 {
		return nil
	}
	tags := make([]string, len(inv.Redact))
	for i := range inv.Redact {
		tags[i] = "secret"
	}
	return tags
}
