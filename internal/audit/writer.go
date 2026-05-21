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

// Writer appends Event records to a daily-rotated JSONL file in dir.
// All writes are serialized; Write is safe for concurrent use.
type Writer struct {
	mu       sync.Mutex
	dir      string
	redactor *Redactor
	day      string // current UTC date suffix, e.g. "2026-05-21"
	f        *os.File
}

// NewWriter creates a Writer rooted at dir. The directory is created if missing.
// redactor masks declared secret substrings from probe_payload before each line
// is serialized. Pass NewRedactor(nil) for a no-op redactor.
func NewWriter(dir string, redactor *Redactor) (*Writer, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("audit: mkdir %s: %w", dir, err)
	}
	return &Writer{dir: dir, redactor: redactor}, nil
}

// Write serializes ev to one JSONL line in the day-rotated file.
// secrets in ProbePayload are redacted before serialization.
func (w *Writer) Write(ev Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if ev.Chronon.IsZero() {
		ev.Chronon = time.Now().UTC()
	}

	day := ev.Chronon.UTC().Format("2006-01-02")
	if day != w.day || w.f == nil {
		if w.f != nil {
			_ = w.f.Close()
		}
		path := filepath.Join(w.dir, day+".jsonl")
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("audit: open %s: %w", path, err)
		}
		w.f = f
		w.day = day
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
