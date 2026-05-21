package brew

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/shell"
)

// localToRemoteSeam bundles the host-touching dependencies that
// UpdateLocalToRemote needs so tests can substitute pure-Go fakes for
// `brew bundle dump` and the y/n prompt. The zero value is unusable;
// production code uses defaultLocalToRemoteSeam.
type localToRemoteSeam struct {
	detect func() (path string, installed bool)
	run    func(rt *config.Runtime, name string, args ...string) (string, error)
	prompt func(rt *config.Runtime, msg string) (bool, error)
}

func defaultLocalToRemoteSeam() localToRemoteSeam {
	return localToRemoteSeam{
		detect: Detect,
		run:    shell.Run,
		prompt: func(rt *config.Runtime, msg string) (bool, error) {
			return shell.Prompt(rt, os.Stdin, msg)
		},
	}
}

// UpdateLocalToRemote dumps the live brew state, diffs it against the
// repo Brewfile, prompts for confirmation, and rewrites <repo>/Brewfile
// on confirm. Refuses when no repo is configured (embed-fallback mode);
// the embedded Brewfile is read-only.
//
// The flow is:
//  1. Resolve the repo path; refuse if none configured.
//  2. Detect brew; refuse if not installed.
//  3. Capture `<brew> bundle dump --describe --file=- --force` stdout
//     as the "local" Brewfile.
//  4. Read the repo Brewfile as the "remote" Brewfile (an absent file
//     is treated as empty so the very first sync from a populated Mac
//     to a fresh repo writes the file cleanly).
//  5. Diff and render the added/removed entries to stderr.
//  6. Prompt the user; rt.Yes short-circuits the prompt.
//  7. Build a canonical merge (Unchanged from local for Extras drift +
//     Added) and atomically rewrite the repo Brewfile via a tempfile
//     rename. Honors rt.DryRun.
func UpdateLocalToRemote(rt *config.Runtime) error {
	return updateLocalToRemote(defaultLocalToRemoteSeam(), rt)
}

func updateLocalToRemote(s localToRemoteSeam, rt *config.Runtime) error {
	repoPath, _, err := rt.ResolveRepoPath()
	if err != nil {
		return err
	}
	if repoPath == "" {
		return errors.New("update local-to-remote: no repo configured; clone the macheim repo and set --repo or MACHEIM_REPO first")
	}

	brewPath, installed := s.detect()
	if !installed {
		return errors.New("brew not installed; run `macheim brew install` first")
	}

	local, err := dumpLocalBrewfile(s, rt, brewPath)
	if err != nil {
		return err
	}

	remote, err := readRemoteBrewfile(repoPath)
	if err != nil {
		return err
	}

	result := Diff(local, remote)
	renderBrewDrift(result)
	if len(result.Added) == 0 && len(result.Removed) == 0 {
		return nil
	}

	confirmed, err := s.prompt(rt, fmt.Sprintf("Apply these changes to %s?", filepath.Join(repoPath, "Brewfile")))
	if err != nil {
		return err
	}
	if !confirmed {
		_, _ = fmt.Fprintln(os.Stderr, "brew: skipped")
		return nil
	}

	merged := mergeBrewfile(local, result)
	return writeRepoBrewfile(rt, repoPath, merged)
}

// dumpLocalBrewfile runs `brew bundle dump --describe --file=- --force`
// and parses the output. The --force flag is harmless when writing to
// stdout but matches the canonical recipe so the command works whether
// the user invokes it directly or via macheim.
func dumpLocalBrewfile(s localToRemoteSeam, rt *config.Runtime, brewPath string) (Brewfile, error) {
	// We deliberately do not propagate rt.Verbose into a custom message
	// here — shell.Run already echoes the command when verbose, and the
	// `bundle dump` output itself is the user-facing artifact.
	dump, err := s.run(rt, brewPath, "bundle", "dump", "--describe", "--file=-", "--force")
	if err != nil {
		return Brewfile{}, fmt.Errorf("brew bundle dump: %w", err)
	}
	bf, err := Parse(strings.NewReader(dump))
	if err != nil {
		return Brewfile{}, fmt.Errorf("parse `brew bundle dump` output: %w", err)
	}
	return bf, nil
}

// readRemoteBrewfile reads and parses the repo's Brewfile. A missing
// file is not an error: it produces an empty Brewfile so the first
// local-to-remote sync against a fresh repo writes the file cleanly.
func readRemoteBrewfile(repoPath string) (Brewfile, error) {
	path := filepath.Join(repoPath, "Brewfile")
	f, err := os.Open(path) //nolint:gosec // repoPath comes from validated config
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Brewfile{}, nil
		}
		return Brewfile{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	bf, err := Parse(f)
	if err != nil {
		return Brewfile{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return bf, nil
}

// renderBrewDrift prints a human-readable summary of the diff to stderr.
// An empty diff prints "brew: no drift" so the user sees a confirming
// line rather than silence.
func renderBrewDrift(result DiffResult) {
	if len(result.Added) == 0 && len(result.Removed) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "brew: no drift")
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, "Brew drift:")
	for _, e := range result.Added {
		_, _ = fmt.Fprintf(os.Stderr, "  + %s %s\n", e.Kind.String(), e.Name)
	}
	for _, e := range result.Removed {
		_, _ = fmt.Fprintf(os.Stderr, "  - %s %s\n", e.Kind.String(), e.Name)
	}
}

// mergeBrewfile builds the canonical Brewfile that local-to-remote will
// write back to the repo. The output is the local entries grouped by
// Kind (tap, brew, cask, mas) and sorted by Name within each group.
// Sorting produces a stable file that is easy to review in `git diff`
// — the alternative (preserve dump order) flips lines on every sync
// because `brew bundle dump` does not promise a stable order.
//
// `Removed` is consulted only as a sanity gate: every Removed entry
// must be absent from `local`, which Diff already guarantees. We
// nonetheless filter defensively in case a future caller threads
// the merge through a custom DiffResult.
func mergeBrewfile(local Brewfile, result DiffResult) Brewfile {
	removed := make(map[entryKey]struct{}, len(result.Removed))
	for _, e := range result.Removed {
		removed[e.key()] = struct{}{}
	}
	var kept []Entry
	for _, e := range local.Entries {
		if _, drop := removed[e.key()]; drop {
			continue
		}
		kept = append(kept, e)
	}
	sort.SliceStable(kept, func(i, j int) bool {
		if kept[i].Kind != kept[j].Kind {
			return kept[i].Kind < kept[j].Kind
		}
		return kept[i].Name < kept[j].Name
	})
	return Brewfile{Entries: kept}
}

// writeRepoBrewfile rewrites <repo>/Brewfile atomically via a tempfile
// rename so a crash mid-write never leaves a partial file. Honors
// rt.DryRun: prints "[dry-run] would rewrite <path>" to stderr and
// returns nil without touching the filesystem.
func writeRepoBrewfile(rt *config.Runtime, repoPath string, bf Brewfile) error {
	path := filepath.Join(repoPath, "Brewfile")
	if rt != nil && rt.DryRun {
		_, _ = fmt.Fprintf(os.Stderr, "[dry-run] would rewrite %s\n", path)
		return nil
	}

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".Brewfile.tmp-*")
	if err != nil {
		return fmt.Errorf("create tempfile in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	// Best-effort cleanup if a later step fails after CreateTemp succeeded.
	cleanup := func() { _ = os.Remove(tmpPath) }

	if err := bf.Write(tmp); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("write Brewfile: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close Brewfile tempfile: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, path, err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "brew: rewrote %s\n", path)
	return nil
}
