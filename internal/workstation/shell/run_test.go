//go:build !windows

package shell

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// captureStderr replaces os.Stderr with a pipe for the duration of fn and
// returns whatever was written. The pipe is drained in a goroutine so a
// flood of bytes from fn cannot deadlock on a full pipe buffer.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stderr = w

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()

	_ = w.Close()
	os.Stderr = orig
	return <-done
}

func TestRun_DryRunSkipsExec(t *testing.T) {
	// /usr/bin/false exits 1 if actually invoked; under --dry-run we must
	// never reach exec, so the returned error must be nil.
	rt := &config.Runtime{DryRun: true}
	stderr := captureStderr(t, func() {
		out, err := Run(rt, "/usr/bin/false")
		if err != nil {
			t.Fatalf("dry-run Run returned err: %v", err)
		}
		if out != "" {
			t.Fatalf("dry-run Run returned non-empty stdout %q", out)
		}
	})
	if !strings.Contains(stderr, "[dry-run] /usr/bin/false") {
		t.Fatalf("stderr missing dry-run prefix: %q", stderr)
	}
}

func TestRun_Captures(t *testing.T) {
	rt := &config.Runtime{Quiet: true}
	out, err := Run(rt, "/bin/echo", "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("captured stdout missing %q: got %q", "hello", out)
	}
}

func TestRun_VerbosePrefixesEcho(t *testing.T) {
	rt := &config.Runtime{Verbose: true, Quiet: true}
	stderr := captureStderr(t, func() {
		if _, err := Run(rt, "/usr/bin/true"); err != nil {
			t.Fatalf("Run: %v", err)
		}
	})
	if !strings.Contains(stderr, "$ /usr/bin/true") {
		t.Fatalf("stderr missing verbose echo: %q", stderr)
	}
}

func TestRun_NilRuntimeIsZeroValue(t *testing.T) {
	out, err := Run(nil, "/bin/echo", "ok")
	if err != nil {
		t.Fatalf("Run with nil rt: %v", err)
	}
	if !strings.Contains(out, "ok") {
		t.Fatalf("captured stdout missing %q: got %q", "ok", out)
	}
}