// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriter_AppendsJSONL(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWriter(dir, NewRedactor([]string{"secret"}))
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	ev := Event{
		Chronon: time.Date(2026, 5, 21, 14, 30, 22, int(481*time.Millisecond), time.UTC),
		Trace:   "01HVQK",
		Cwd:     "/work",
		Actor:   ActorMacforge,
		Kind:    KindInvocationAttempt,
		Probe:   "codesign",
		ProbePayload: "--password secret --sign id",
	}
	if err := w.Write(ev); err != nil {
		t.Fatalf("Write: %v", err)
	}

	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f, _ := os.Open(files[0])
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Scan()
	line := sc.Text()

	if strings.Contains(line, "secret") {
		t.Fatalf("audit line contains unredacted secret: %q", line)
	}
	if !strings.Contains(line, "[REDACTED]") {
		t.Fatalf("audit line missing [REDACTED] marker: %q", line)
	}

	var got Event
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("audit line is not valid JSON: %v", err)
	}
	if got.Trace != "01HVQK" {
		t.Fatalf("Trace = %q, want %q", got.Trace, "01HVQK")
	}
	if got.Kind != KindInvocationAttempt {
		t.Fatalf("Kind = %q, want %q", got.Kind, KindInvocationAttempt)
	}
}

func TestWriter_DailyRotation(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWriter(dir, NewRedactor(nil))
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	defer w.Close()

	day1 := time.Date(2026, 5, 21, 23, 59, 0, 0, time.UTC)
	day2 := time.Date(2026, 5, 22, 0, 0, 1, 0, time.UTC)

	if err := w.Write(Event{Chronon: day1, Trace: "T1", Actor: ActorMacforge, Kind: KindSignal}); err != nil {
		t.Fatalf("write day1: %v", err)
	}
	if err := w.Write(Event{Chronon: day2, Trace: "T1", Actor: ActorMacforge, Kind: KindSignal}); err != nil {
		t.Fatalf("write day2: %v", err)
	}

	files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
	if len(files) != 2 {
		t.Fatalf("expected 2 daily files, got %d: %v", len(files), files)
	}
}

func TestNewTraceID_Format(t *testing.T) {
	id := NewTraceID()
	if len(id) != 26 {
		t.Fatalf("ULID length = %d, want 26", len(id))
	}
}
