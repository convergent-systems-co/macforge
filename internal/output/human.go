// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package output

import (
	"fmt"
	"io"
	"strings"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
)

func renderHumanSuccess(w io.Writer, r Outputter, nocolor bool) error {
	prefix := "✓"
	if !nocolor {
		prefix = ansiGreen + ansiBold + "✓" + ansiReset
	}
	if _, err := fmt.Fprintf(w, "%s %s\n", prefix, r.SchemaName()); err != nil {
		return err
	}
	for _, line := range r.HumanLines() {
		if _, err := fmt.Fprintln(w, "  "+line); err != nil {
			return err
		}
	}
	return nil
}

func renderHumanFailure(w io.Writer, err *mferrors.Error, nocolor bool) error {
	prefix := "✗ " + err.Code
	if !nocolor {
		prefix = ansiRed + ansiBold + "✗ " + err.Code + ansiReset
	}
	header := fmt.Sprintf("%s — %s\n", prefix, err.Msg)
	if _, werr := fmt.Fprint(w, header); werr != nil {
		return werr
	}
	if _, werr := fmt.Fprintf(w, "  op:   %s\n", err.Op); werr != nil {
		return werr
	}
	if err.Hint != "" {
		if _, werr := fmt.Fprintf(w, "  hint: %s\n", err.Hint); werr != nil {
			return werr
		}
	}
	if len(err.Details) > 0 {
		if _, werr := fmt.Fprintln(w, "  details:"); werr != nil {
			return werr
		}
		for k, v := range err.Details {
			if _, werr := fmt.Fprintf(w, "    %s: %v\n", k, v); werr != nil {
				return werr
			}
		}
	}
	if err.Cause != nil {
		if _, werr := fmt.Fprintf(w, "  cause: %s\n", strings.TrimSpace(err.Cause.Error())); werr != nil {
			return werr
		}
	}
	return nil
}
