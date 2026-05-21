// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/convergent-systems-co/macforge/internal/audit"
	"github.com/convergent-systems-co/macforge/internal/config"
	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
	"github.com/convergent-systems-co/macforge/internal/output"
)

// cliRuntime carries the per-invocation singletons: trace ID, audit writer,
// output renderer, and the resolved config (when loaded).
type cliRuntime struct {
	trace    string
	cfg      *config.Config
	audit    *audit.Writer
	renderer *output.Renderer
}

// newRuntime builds a runtime for the current invocation. command is the
// verb name (used in the JSON envelope's "command" field). loadConfig is
// false for commands like `init` that don't yet have a macforge.yaml.
func newRuntime(command string, loadConfig bool) (*cliRuntime, error) {
	rt := &cliRuntime{trace: audit.NewTraceID()}

	// Per ADR-0016: per-invocation file at ~/.macforge/audit/<UTC>-<trace>.jsonl
	now := time.Now().UTC()
	auditFile := filepath.Join(
		config.AuditDir(),
		now.Format("2006-01-02T15-04-05Z")+"-"+rt.trace+".jsonl",
	)
	w, err := audit.NewWriter(auditFile, audit.NewRedactor(nil))
	if err != nil {
		return nil, mferrors.NewAudit(mferrors.CodeAuditWriteFail, "runtime.newRuntime",
			"failed to open audit writer", mferrors.WithCause(err))
	}
	rt.audit = w

	mode := output.ParseMode(resolveOutputMode())
	rt.renderer = output.NewRenderer(mode, stdoutForRenderer, rt.trace, command, gflags.noColor || os.Getenv("NO_COLOR") != "")

	if loadConfig {
		cfg, err := config.Load(config.LoadOptions{GlobalPath: gflags.configPath})
		if err != nil {
			return rt, err
		}
		rt.cfg = cfg
	}
	return rt, nil
}

func resolveOutputMode() string {
	if gflags.output != "" {
		return gflags.output
	}
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return "json"
	}
	return "human"
}

// emit either renders success or failure and returns the original error.
// CLI commands defer this to make the runtime present even when the verb
// short-circuits.
func (rt *cliRuntime) emit(out output.Outputter, runErr error) error {
	defer rt.audit.Close()
	if runErr != nil {
		var mfErr *mferrors.Error
		if !asError(runErr, &mfErr) {
			mfErr = mferrors.NewConfig(mferrors.CodeConfigInvalid, "cli", runErr.Error())
		}
		_ = rt.renderer.Failure(mfErr)
		return mfErr
	}
	return rt.renderer.Success(out)
}

// asError is errors.As specialized to *mferrors.Error.
func asError(err error, target **mferrors.Error) bool {
	for cur := err; cur != nil; {
		if mf, ok := cur.(*mferrors.Error); ok {
			*target = mf
			return true
		}
		type unwrapper interface{ Unwrap() error }
		u, ok := cur.(unwrapper)
		if !ok {
			return false
		}
		cur = u.Unwrap()
	}
	return false
}

// stdoutForRenderer is a hook for tests to swap stdout.
var stdoutForRenderer io.Writer = os.Stdout

// macforgeYAMLPath returns the path init should write to: the global config
// at ${XDG_CONFIG_HOME:-$HOME/.config}/macforge/macforge.yaml. Per ADR-0015,
// init writes ONLY the global file; project-local overrides are user-authored.
func macforgeYAMLPath() string {
	return config.ConfigPath()
}
