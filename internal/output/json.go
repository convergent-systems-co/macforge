// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package output

import (
	"encoding/json"
	"io"

	mferrors "github.com/convergent-systems-co/macforge/internal/errors"
)

func renderJSONSuccess(w io.Writer, trace, command string, r Outputter) error {
	env := successEnvelope{
		OK:      true,
		Schema:  r.SchemaName(),
		Trace:   trace,
		Command: command,
		Result:  r,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.Write(b)
	return err
}

func renderJSONFailure(w io.Writer, trace, command string, mfErr *mferrors.Error) error {
	env := errorEnvelope{
		OK:      false,
		Schema:  "macforge.v1.error",
		Trace:   trace,
		Command: command,
		Code:    mfErr.Code,
		Op:      mfErr.Op,
		Message: mfErr.Msg,
		Hint:    mfErr.Hint,
		Details: mfErr.Details,
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = w.Write(b)
	return err
}
