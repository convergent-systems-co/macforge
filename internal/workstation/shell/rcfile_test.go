package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

func TestAppendIfMissing_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	line := `eval "$(brew shellenv)"`

	if err := AppendIfMissing(&config.Runtime{}, rc, line); err != nil {
		t.Fatalf("AppendIfMissing: %v", err)
	}

	got, err := os.ReadFile(rc)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != line+"\n" {
		t.Fatalf("file content: got %q want %q", string(got), line+"\n")
	}
}

func TestAppendIfMissing_IsIdempotent(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	line := `eval "$(brew shellenv)"`

	for i := 0; i < 3; i++ {
		if err := AppendIfMissing(&config.Runtime{}, rc, line); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}

	got, err := os.ReadFile(rc)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if n := strings.Count(string(got), line); n != 1 {
		t.Fatalf("line appears %d times, want 1: file is %q", n, string(got))
	}
}

func TestAppendIfMissing_PreservesOtherLines(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	existing := "export FOO=bar\nexport BAZ=qux\n"
	if err := os.WriteFile(rc, []byte(existing), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	newLine := `eval "$(brew shellenv)"`

	if err := AppendIfMissing(&config.Runtime{}, rc, newLine); err != nil {
		t.Fatalf("AppendIfMissing: %v", err)
	}

	got, err := os.ReadFile(rc)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	want := existing + newLine + "\n"
	if string(got) != want {
		t.Fatalf("file content: got %q want %q", string(got), want)
	}
}

func TestAppendIfMissing_DryRunNoFs(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	line := `eval "$(brew shellenv)"`

	stderr := captureStderr(t, func() {
		if err := AppendIfMissing(&config.Runtime{DryRun: true}, rc, line); err != nil {
			t.Fatalf("AppendIfMissing: %v", err)
		}
	})

	if _, err := os.Stat(rc); !os.IsNotExist(err) {
		t.Fatalf("dry-run created file: stat err = %v", err)
	}
	if !strings.Contains(stderr, "[dry-run]") || !strings.Contains(stderr, rc) {
		t.Fatalf("dry-run stderr missing intent: %q", stderr)
	}
}

func TestAppendIfMissing_DetectsLineMidFile(t *testing.T) {
	// Idempotency must catch the line wherever it sits in the file, not
	// only at the tail. Seed the rc with the target line bracketed by
	// other content; AppendIfMissing must no-op.
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")
	line := `eval "$(brew shellenv)"`
	seed := "export FOO=bar\n" + line + "\nexport BAZ=qux\n"
	if err := os.WriteFile(rc, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := AppendIfMissing(&config.Runtime{}, rc, line); err != nil {
		t.Fatalf("AppendIfMissing: %v", err)
	}

	got, err := os.ReadFile(rc)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != seed {
		t.Fatalf("file mutated: got %q want %q", string(got), seed)
	}
}
