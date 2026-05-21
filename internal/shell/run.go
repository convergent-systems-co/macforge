// Package shell provides the operational plumbing that every mutating
// macheim command shares: subprocess execution with live output streaming,
// y/n prompts that honor --yes, shell detection ($SHELL → rc file path),
// and idempotent rc-file line appends. All entry points consult the
// caller's *config.Runtime so --dry-run, --quiet, --verbose, and --yes
// behave identically across commands.
package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/polliard/macheim/internal/config"
)

// Run executes a command, streaming stdout and stderr live, and returns
// the captured stdout. Behavior is selected by the *Runtime flags:
//
//   - rt.Verbose: echo "$ <cmd>" to os.Stderr before running (also when
//     DryRun is set, so verbose dry-runs show the intended invocation).
//   - rt.DryRun:  print "[dry-run] <cmd>" to os.Stderr and return ("", nil)
//     without invoking exec.
//   - rt.Quiet:   capture stdout into the returned string, but suppress
//     the live stream to the user's terminal. stderr still passes through
//     so error output from the child process is never lost.
//
// stderr from the child process always streams live to os.Stderr — even
// under --quiet — so the user sees diagnostics from misbehaving tools.
// A nil rt is treated as the zero value (all flags false): real exec,
// live stream, no dry-run.
func Run(rt *config.Runtime, name string, args ...string) (string, error) {
	line := commandLine(name, args)

	dryRun, verbose, quiet := false, false, false
	if rt != nil {
		dryRun, verbose, quiet = rt.DryRun, rt.Verbose, rt.Quiet
	}

	if verbose {
		_, _ = fmt.Fprintf(os.Stderr, "$ %s\n", line)
	}
	if dryRun {
		_, _ = fmt.Fprintf(os.Stderr, "[dry-run] %s\n", line)
		return "", nil
	}

	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr

	var buf bytes.Buffer
	if quiet {
		cmd.Stdout = &buf
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	}

	err := cmd.Run()
	return buf.String(), err
}

// commandLine renders a command + args for log echoing. It is intentionally
// simple — no shell quoting — because the output is informational, not
// designed to be copy-pasted into a different shell.
func commandLine(name string, args []string) string {
	if len(args) == 0 {
		return name
	}
	return name + " " + strings.Join(args, " ")
}
