package brew

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
	"github.com/convergent-systems-co/macforge/internal/workstation/embedded"
	"github.com/convergent-systems-co/macforge/internal/workstation/shell"
)

// embedBrewfilePath names the embedded Brewfile inside the embedded.Configs
// FS. Centralized so the seam fallback and the production path agree.
const embedBrewfilePath = "configs/Brewfile"

// bundleSeam bundles the host-touching dependencies Apply needs. The
// zero value is unusable; use defaultBundleSeam. Same shape as the
// installSeam pattern in install.go so the two surfaces test alike.
type bundleSeam struct {
	detect    func() (path string, installed bool)
	resolve   func(rt *config.Runtime) (path, source string, err error)
	stat      func(string) (os.FileInfo, error)
	embedRead func() ([]byte, error)
	tempWrite func(content []byte) (path string, cleanup func(), err error)
	run       func(rt *config.Runtime, name string, args ...string) (string, error)
	stderr    func(format string, args ...any)
}

func defaultBundleSeam() bundleSeam {
	return bundleSeam{
		detect:    Detect,
		resolve:   func(rt *config.Runtime) (string, string, error) { return rt.ResolveRepoPath() },
		stat:      os.Stat,
		embedRead: func() ([]byte, error) { return embedded.Configs.ReadFile(embedBrewfilePath) },
		tempWrite: func(content []byte) (string, func(), error) {
			f, err := os.CreateTemp("", "macheim-Brewfile-*")
			if err != nil {
				return "", func() {}, fmt.Errorf("create temp Brewfile: %w", err)
			}
			path := f.Name()
			cleanup := func() { _ = os.Remove(path) }
			if _, err := f.Write(content); err != nil {
				_ = f.Close()
				cleanup()
				return "", func() {}, fmt.Errorf("write temp Brewfile: %w", err)
			}
			if err := f.Close(); err != nil {
				cleanup()
				return "", func() {}, fmt.Errorf("close temp Brewfile: %w", err)
			}
			return path, cleanup, nil
		},
		run: shell.Run,
		stderr: func(format string, args ...any) {
			_, _ = fmt.Fprintf(os.Stderr, format, args...)
		},
	}
}

// Apply runs `brew bundle --file=<Brewfile>` against the configured
// repo's Brewfile, falling back to the embedded snapshot when no repo
// is configured (or when the configured repo has no Brewfile yet).
//
// Honors:
//   - rt.DryRun  — propagated through shell.Run, which echoes the
//     intended command and returns without executing.
//   - rt.Verbose — shell.Run echoes the command; Apply prints one
//     "using <path>" line so the user knows which Brewfile drove the
//     run.
//   - rt.Quiet   — shell.Run suppresses live stdout (still captures
//     it). Apply still prints the one-line summary because that is
//     the user-facing result of the operation.
//
// Returns an error when brew is not installed (with the remediation
// pointing at `macheim brew install`), when repo resolution fails,
// or when the brew bundle subprocess exits non-zero.
func Apply(rt *config.Runtime) error {
	return apply(defaultBundleSeam(), rt)
}

func apply(s bundleSeam, rt *config.Runtime) error {
	brewPath, installed := s.detect()
	if !installed {
		return errors.New("brew is not installed; run `macheim brew install` first")
	}

	brewfilePath, cleanup, err := resolveBrewfile(s, rt)
	if err != nil {
		return err
	}
	defer cleanup()

	s.stderr("brew bundle: using %s\n", brewfilePath)

	out, err := s.run(rt, brewPath, "bundle", "--file="+brewfilePath)
	if err != nil {
		return fmt.Errorf("brew bundle: %w", err)
	}

	if rt != nil && rt.DryRun {
		// shell.Run short-circuited before exec; no output to summarize.
		return nil
	}
	installedN, skippedN := summarize(out)
	s.stderr("brew bundle: %d installed, %d skipped\n", installedN, skippedN)
	return nil
}

// resolveBrewfile picks the on-disk Brewfile path Apply should hand to
// `brew bundle`. Preference order:
//
//  1. <repo>/Brewfile when the repo resolved and the file exists.
//  2. The embedded snapshot, written to an OS temp file so brew bundle
//     (which only accepts a path) can read it.
//
// The returned cleanup is always non-nil and safe to defer; for the
// repo path it is a no-op, for the embed path it removes the temp file.
func resolveBrewfile(s bundleSeam, rt *config.Runtime) (path string, cleanup func(), err error) {
	repoPath, _, err := s.resolve(rt)
	if err != nil {
		return "", func() {}, fmt.Errorf("brew bundle: resolve repo: %w", err)
	}
	if repoPath != "" {
		candidate := filepath.Join(repoPath, "Brewfile")
		if _, statErr := s.stat(candidate); statErr == nil {
			return candidate, func() {}, nil
		}
	}

	content, err := s.embedRead()
	if err != nil {
		return "", func() {}, fmt.Errorf("brew bundle: read embedded Brewfile: %w", err)
	}
	tempPath, tempCleanup, err := s.tempWrite(content)
	if err != nil {
		return "", func() {}, fmt.Errorf("brew bundle: stage embedded Brewfile: %w", err)
	}
	return tempPath, tempCleanup, nil
}

// summarize reads brew bundle's stdout and returns (installed, skipped)
// counts. brew bundle prints one status line per entry; the surface we
// care about is:
//
//	Installing <name>      — newly installed this run
//	Using <name>           — already present; counted as skipped
//	Skipping <name>        — explicitly skipped (e.g. unsupported)
//
// Other lines (summary footers, warnings) are ignored. The match is on
// the leading word of the trimmed line so brew's optional emoji or color
// prefixes — which arrive only when stdout is a TTY, never under our
// capture — don't matter in practice.
func summarize(output string) (installedN, skippedN int) {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		first := trimmed
		if i := strings.IndexByte(trimmed, ' '); i >= 0 {
			first = trimmed[:i]
		}
		switch first {
		case "Installing":
			installedN++
		case "Using", "Skipping":
			skippedN++
		}
	}
	return installedN, skippedN
}
