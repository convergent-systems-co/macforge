// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Writer appends Event records as JSONL to a single file. The file is
// determined at construction time — per ADR-0016 each `macforge`
// invocation writes to its own file at ~/.macforge/audit/<UTC>-<trace>.jsonl.
// No rotation logic here; the runtime computes the filename.
type Writer struct {
	mu       sync.Mutex
	path     string
	redactor *Redactor
	f        *os.File
}

// NewWriter creates a Writer that appends to path. The parent directory
// is created if missing. redactor masks declared secret substrings from
// probe_payload and note before each line is serialized; pass
// NewRedactor(nil) for a no-op redactor.
func NewWriter(path string, redactor *Redactor) (*Writer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("audit: mkdir %s: %w", filepath.Dir(path), err)
	}
	return &Writer{path: path, redactor: redactor}, nil
}

// Write serializes ev to one JSONL line. The file is opened on first
// Write and held open until Close. Safe for concurrent use.
func (w *Writer) Write(ev Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ev.Chronon.IsZero() {
		ev.Chronon = time.Now().UTC()
	}

	if w.f == nil {
		f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("audit: open %s: %w", w.path, err)
		}
		w.f = f
	}

	if w.redactor != nil {
		ev.ProbePayload = w.redactor.Apply(ev.ProbePayload)
		ev.Note = w.redactor.Apply(ev.Note)
	}

	b, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("audit: marshal: %w", err)
	}
	b = append(b, '\n')

	if _, err := w.f.Write(b); err != nil {
		return fmt.Errorf("audit: write: %w", err)
	}
	return nil
}

// Close releases the underlying file.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}

// Path returns the file this Writer is appending to. Useful in the
// result envelope so the caller can surface the audit-file path to
// the operator.
func (w *Writer) Path() string {
	return w.path
}
