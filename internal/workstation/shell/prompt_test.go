//go:build !windows

package shell

import (
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

func TestPrompt_YesShortCircuits(t *testing.T) {
	rt := &config.Runtime{Yes: true}
	// Empty reader: if Yes did not short-circuit, ReadString would hit EOF
	// and return false. A true result proves the short-circuit fired.
	stderr := captureStderr(t, func() {
		ok, err := Prompt(rt, strings.NewReader(""), "proceed?")
		if err != nil {
			t.Fatalf("Prompt: %v", err)
		}
		if !ok {
			t.Fatalf("expected true from --yes short-circuit")
		}
	})
	if stderr != "" {
		t.Fatalf("--yes should not print prompt; got stderr %q", stderr)
	}
}

func TestPrompt_YInput(t *testing.T) {
	ok, err := Prompt(&config.Runtime{}, strings.NewReader("y\n"), "go?")
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if !ok {
		t.Fatalf("expected true for y input")
	}
}

func TestPrompt_YesInput(t *testing.T) {
	ok, err := Prompt(&config.Runtime{}, strings.NewReader("yes\n"), "go?")
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if !ok {
		t.Fatalf("expected true for yes input")
	}
}

func TestPrompt_UppercaseYes(t *testing.T) {
	ok, err := Prompt(&config.Runtime{}, strings.NewReader("YES\n"), "go?")
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if !ok {
		t.Fatalf("expected true for YES input")
	}
}

func TestPrompt_EmptyDefaultsNo(t *testing.T) {
	ok, err := Prompt(&config.Runtime{}, strings.NewReader("\n"), "go?")
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if ok {
		t.Fatalf("expected false for empty input")
	}
}

func TestPrompt_OtherInputNo(t *testing.T) {
	ok, err := Prompt(&config.Runtime{}, strings.NewReader("maybe\n"), "go?")
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if ok {
		t.Fatalf("expected false for non-y input")
	}
}

func TestPrompt_EOFDefaultsNo(t *testing.T) {
	// Closed-stdin scenario: no newline, immediate EOF. Must default to
	// "no" — never interpret a broken pipe as consent.
	ok, err := Prompt(&config.Runtime{}, strings.NewReader(""), "go?")
	if err != nil {
		t.Fatalf("Prompt: %v", err)
	}
	if ok {
		t.Fatalf("expected false on EOF")
	}
}