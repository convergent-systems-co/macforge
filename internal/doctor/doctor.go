// Package doctor implements `macheim doctor`: read-only environment
// sanity-checks with structured per-check results and an aggregated exit
// code. Output rendering lives in render.go; the orchestrator (Run) and
// the check protocol (Check / Result) live here.
package doctor

import (
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

// Result is the structured outcome of a single check.
type Result struct {
	OK          bool
	Probe       string // shown only under --verbose
	Remediation string // shown on failure regardless of --verbose
}

// Check is one named diagnostic. Run is invoked once per `macheim doctor`
// invocation, in the fixed order returned by DefaultChecks.
type Check struct {
	Name string
	Run  func(rt *config.Runtime) Result
}

// seam bundles OS-touching dependencies so tests can substitute fakes.
type seam struct {
	run          func(name string, args ...string) (string, error)
	stat         func(string) (os.FileInfo, error)
	lookEnv      func(string) string
	canWriteDir  func(string) bool
	canWriteFile func(string) bool
	arch         string
	homeDir      string
}

func defaultSeam() seam {
	home, _ := os.UserHomeDir()
	return seam{
		run: func(name string, args ...string) (string, error) {
			out, err := exec.Command(name, args...).Output() //nolint:gosec // intentional; argv is constant in production checks
			return string(out), err
		},
		stat:         os.Stat,
		lookEnv:      os.Getenv,
		canWriteDir:  dirWritable,
		canWriteFile: fileOrParentWritable,
		arch:         runtime.GOARCH,
		homeDir:      home,
	}
}

// DefaultChecks returns the canonical list of checks, in display order.
func DefaultChecks() []Check {
	s := defaultSeam()
	return []Check{
		{Name: "xcode-select", Run: func(rt *config.Runtime) Result { return xcodeCheck(rt, s) }},
		{Name: "brew", Run: func(rt *config.Runtime) Result { return brewCheck(rt, s) }},
		{Name: "repo", Run: func(rt *config.Runtime) Result { return repoCheck(rt, s) }},
		{Name: "config-dir", Run: func(rt *config.Runtime) Result { return configDirCheck(rt, s) }},
		{Name: "shell-rc", Run: func(rt *config.Runtime) Result { return shellRCCheck(rt, s) }},
	}
}

// Run executes every check in order, accumulates failures, and returns
// nil on all-pass or a cli.ExitCoder with exit code 1 on any failure.
// Rendering is added in a follow-up commit; the writer is accepted here
// to keep the public signature stable but is unused for now.
func Run(rt *config.Runtime, w io.Writer) error {
	_ = w // render lands in render.go in the next commit
	failed := 0
	for _, c := range DefaultChecks() {
		if !c.Run(rt).OK {
			failed++
		}
	}
	if failed > 0 {
		return cli.Exit("", 1)
	}
	return nil
}
