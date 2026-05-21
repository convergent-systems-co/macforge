// Package brew owns the Homebrew lifecycle commands: arch-aware detection
// of the canonical brew binary, installing brew from a pinned embedded
// snapshot of the official installer, and (in later sub-issues) brew
// bundle and dump-vs-repo diffs.
//
// Detection and install are factored behind a small seam struct so the
// public entry points (Detect, Version, Install) can be exercised with
// fakes — the same pattern used by internal/doctor.
package brew

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// detectSeam bundles the host-touching dependencies that Detect and
// Version need. The zero value is unusable; use defaultDetectSeam.
type detectSeam struct {
	arch string
	stat func(string) (os.FileInfo, error)
	run  func(name string, args ...string) (string, error)
}

func defaultDetectSeam() detectSeam {
	return detectSeam{
		arch: runtime.GOARCH,
		stat: os.Stat,
		run: func(name string, args ...string) (string, error) {
			out, err := exec.Command(name, args...).Output() //nolint:gosec // argv is constant from caller
			return string(out), err
		},
	}
}

// Detect returns the canonical brew path for the host architecture and
// reports whether the binary exists at that path.
//
//   - arm64 (Apple Silicon) → /opt/homebrew/bin/brew
//   - amd64 (Intel)         → /usr/local/bin/brew
//   - anything else         → ("", false)
//
// Presence is probed with os.Stat. This mirrors internal/doctor.brewCheck
// so the two surfaces agree on what "brew is installed" means.
func Detect() (path string, installed bool) {
	return detect(defaultDetectSeam())
}

func detect(s detectSeam) (path string, installed bool) {
	switch s.arch {
	case "arm64":
		path = "/opt/homebrew/bin/brew"
	case "amd64":
		path = "/usr/local/bin/brew"
	default:
		return "", false
	}
	if _, err := s.stat(path); err != nil {
		return path, false
	}
	return path, true
}

// Version runs `<brewPath> --version` and returns the version token from
// the first line of output. Homebrew prints lines like:
//
//	Homebrew 4.3.5
//	Homebrew/homebrew-core (git revision abc; last commit 2026-05-01)
//
// so the parse is "split first line on whitespace, take the last word."
// Returns ("", err) if the exec itself fails.
func Version(brewPath string) (string, error) {
	return version(defaultDetectSeam(), brewPath)
}

func version(s detectSeam, brewPath string) (string, error) {
	out, err := s.run(brewPath, "--version")
	if err != nil {
		return "", fmt.Errorf("run %s --version: %w", brewPath, err)
	}
	first := out
	if i := strings.IndexByte(out, '\n'); i >= 0 {
		first = out[:i]
	}
	first = strings.TrimSpace(first)
	if first == "" {
		return "", nil
	}
	fields := strings.Fields(first)
	return fields[len(fields)-1], nil
}
