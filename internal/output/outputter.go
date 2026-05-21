// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package output

import (
	"io"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

// Outputter is implemented by every command result type. SchemaName is the
// stable versioned identifier used in the JSON envelope's "schema" field.
// HumanLines returns the human-rendered body, one entry per line, ANSI-free.
type Outputter interface {
	SchemaName() string
	HumanLines() []string
}

// Renderer dispatches Outputter results to the human or JSON backend.
type Renderer struct {
	mode    Mode
	w       io.Writer
	trace   string
	command string
	nocolor bool
}

// NewRenderer returns a Renderer that writes to w. trace and command are
// stamped into every envelope. nocolor disables ANSI in human mode.
func NewRenderer(mode Mode, w io.Writer, trace, command string, nocolor bool) *Renderer {
	return &Renderer{mode: mode, w: w, trace: trace, command: command, nocolor: nocolor}
}

// Success renders a successful result.
func (r *Renderer) Success(out Outputter) error {
	if r.mode == ModeJSON {
		return renderJSONSuccess(r.w, r.trace, r.command, out)
	}
	return renderHumanSuccess(r.w, out, r.nocolor)
}

// Failure renders an error.
func (r *Renderer) Failure(err *mferrors.Error) error {
	if r.mode == ModeJSON {
		return renderJSONFailure(r.w, r.trace, r.command, err)
	}
	return renderHumanFailure(r.w, err, r.nocolor)
}
