package brew

import (
	"fmt"
	"os"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
	"github.com/convergent-systems-co/macforge/internal/workstation/embedded"
	"github.com/convergent-systems-co/macforge/internal/workstation/shell"
)

// shellenvLine is the rc-file line that exposes brew on PATH for future
// shells. AppendIfMissing's exact-line idempotency means re-running
// Install never writes a duplicate.
const shellenvLine = `eval "$(brew shellenv)"`

// installSeam bundles the dependencies Install touches so tests can
// substitute fakes for Detect, the embedded installer script, shell
// command exec, shell detection, the rc-file append, and Version
// verification. The zero value is unusable; use defaultInstallSeam.
type installSeam struct {
	detect          func() (path string, installed bool)
	scripts         func() ([]byte, error)
	tempFile        func(content []byte) (path string, cleanup func(), err error)
	run             func(rt *config.Runtime, name string, args ...string) (string, error)
	shellDetect     func() (shellName, rcPath string, err error)
	appendIfMissing func(rt *config.Runtime, rcPath, line string) error
	version         func(brewPath string) (string, error)
	stderr          func(format string, args ...any)
}

func defaultInstallSeam() installSeam {
	return installSeam{
		detect:  Detect,
		scripts: func() ([]byte, error) { return embedded.Scripts.ReadFile("scripts/install-brew.sh") },
		tempFile: func(content []byte) (string, func(), error) {
			f, err := os.CreateTemp("", "macheim-install-brew-*.sh")
			if err != nil {
				return "", func() {}, fmt.Errorf("create temp file: %w", err)
			}
			path := f.Name()
			cleanup := func() { _ = os.Remove(path) }
			if _, err := f.Write(content); err != nil {
				_ = f.Close()
				cleanup()
				return "", func() {}, fmt.Errorf("write temp file: %w", err)
			}
			if err := f.Close(); err != nil {
				cleanup()
				return "", func() {}, fmt.Errorf("close temp file: %w", err)
			}
			return path, cleanup, nil
		},
		run:             shell.Run,
		shellDetect:     shell.Detect,
		appendIfMissing: shell.AppendIfMissing,
		version:         Version,
		stderr: func(format string, args ...any) {
			_, _ = fmt.Fprintf(os.Stderr, format, args...)
		},
	}
}

// Install brings Homebrew up to a known-good state on the host:
//
//  1. Detect whether brew is already installed at the canonical
//     arch-appropriate path. If so, skip the installer and proceed to
//     the shellenv step — the binary may have been installed outside
//     macheim without ever wiring brew into the user's rc file.
//  2. Extract the pinned embedded install-brew.sh to a temp file and
//     run it via bash with NONINTERACTIVE=1 (skips the official
//     installer's "press RETURN to continue" prompt; sudo is still
//     required for a first-time install).
//  3. Re-detect after install. If still missing, return an error
//     pointing at the canonical path the installer should have populated.
//  4. Detect the user's shell rc file and append `eval "$(brew shellenv)"`
//     idempotently. Unknown shells are warned-and-skipped, not fatal —
//     the install itself succeeded.
//  5. Run `brew --version` and print the result to os.Stderr. A failure
//     here is informational; the install succeeded.
//
// All shell.Run calls honor rt.DryRun automatically, so --dry-run prints
// what would run without touching disk or executing the installer.
func Install(rt *config.Runtime) error {
	return install(defaultInstallSeam(), rt)
}

func install(s installSeam, rt *config.Runtime) error {
	path, installed := s.detect()
	if path == "" {
		return fmt.Errorf("brew install: unsupported architecture; macheim runs on arm64 (Apple Silicon) and amd64 (Intel) only")
	}

	if installed {
		s.stderr("brew: already installed at %s\n", path)
	} else {
		newPath, err := runInstaller(s, rt, path)
		if err != nil {
			return err
		}
		path = newPath
	}

	if err := writeShellenv(s, rt); err != nil {
		return err
	}
	verifyInstall(s, rt, path)
	return nil
}

// runInstaller extracts the embedded installer to a temp file, runs it
// via bash, and re-detects post-install. Under --dry-run the re-detect
// is skipped (nothing actually ran).
func runInstaller(s installSeam, rt *config.Runtime, path string) (string, error) {
	script, err := s.scripts()
	if err != nil {
		return "", fmt.Errorf("brew install: read embedded installer: %w", err)
	}
	tempPath, cleanup, err := s.tempFile(script)
	if err != nil {
		return "", fmt.Errorf("brew install: stage installer: %w", err)
	}
	defer cleanup()

	// NONINTERACTIVE=1 makes the Homebrew installer skip its
	// "press RETURN" prompt when stdin is a TTY; sudo is still
	// required on a first-time install.
	if _, err := s.run(rt, "bash", "-c", "NONINTERACTIVE=1 "+tempPath); err != nil {
		return "", fmt.Errorf("brew install: run installer: %w", err)
	}

	if rt != nil && rt.DryRun {
		return path, nil
	}
	newPath, installed := s.detect()
	if !installed {
		return "", fmt.Errorf("brew install: installer ran but brew not found at %s; check the output above", newPath)
	}
	return newPath, nil
}

// writeShellenv runs the idempotent rc-file append. An unknown shell is
// warned-and-skipped (the install itself succeeded).
func writeShellenv(s installSeam, rt *config.Runtime) error {
	_, rcPath, err := s.shellDetect()
	if err != nil {
		s.stderr("brew: install succeeded but could not detect shell rc file: %v\n", err)
		s.stderr("brew: add this line to your shell rc manually: %s\n", shellenvLine)
		return nil
	}
	if err := s.appendIfMissing(rt, rcPath, shellenvLine); err != nil {
		return fmt.Errorf("brew install: append shellenv to %s: %w", rcPath, err)
	}
	return nil
}

// verifyInstall runs `brew --version` and prints the result. Skipped
// under --dry-run; failures downgrade to a warning since the install
// itself succeeded.
func verifyInstall(s installSeam, rt *config.Runtime, path string) {
	if rt != nil && rt.DryRun {
		return
	}
	v, err := s.version(path)
	if err != nil {
		s.stderr("brew: install succeeded but `brew --version` failed: %v\n", err)
		return
	}
	s.stderr("brew: installed at %s (%s)\n", path, v)
}
