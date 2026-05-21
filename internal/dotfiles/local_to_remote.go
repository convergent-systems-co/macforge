package dotfiles

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/shell"
)

// LocalToRemoteResult summarizes one UpdateLocalToRemote run.
type LocalToRemoteResult struct {
	// Copied is the list of rel paths that were copied from $HOME back
	// into the repo (or, under --dry-run, that would have been copied).
	Copied []string
	// BackedUp is the list of rel paths whose pre-existing repo file
	// was backed up before overwrite.
	BackedUp []string
	// BackupDir is the absolute path of the timestamped backup directory
	// at <repo>/.macheim-repo-backups/<ISO>. Empty when nothing required
	// backing up.
	BackupDir string
}

// localToRemoteSeam isolates the y/n prompt so tests can answer
// deterministically. The zero value is unusable; production uses
// defaultLocalToRemoteSeam.
type localToRemoteSeam struct {
	prompt func(rt *config.Runtime, msg string) (bool, error)
}

func defaultLocalToRemoteSeam() localToRemoteSeam {
	return localToRemoteSeam{
		prompt: func(rt *config.Runtime, msg string) (bool, error) {
			return shell.Prompt(rt, os.Stdin, msg)
		},
	}
}

// UpdateLocalToRemote walks the repo's dotfiles tree and copies any
// $HOME file that has drifted from its repo counterpart back into the
// repo. The reverse direction of Apply: this is what populates the
// repo when the user has edited a dotfile in place on the Mac.
//
// Out of scope (deliberately): adding files to the repo that the repo
// has never tracked. The repo's dotfiles/ tree is the canonical set
// of tracked paths; "I changed a new file in $HOME, add it to the repo"
// is a separate workflow that requires the user to decide which $HOME
// paths macheim should care about — GOALS.md does not specify it.
//
// Refuses in embed-fallback mode (no repo configured). Pre-existing
// repo files are backed up to <repo>/.macheim-repo-backups/<UTC-ISO>/
// before overwrite so the user can recover from a bad sync.
func UpdateLocalToRemote(rt *config.Runtime) (LocalToRemoteResult, error) {
	return updateLocalToRemote(defaultLocalToRemoteSeam(), rt)
}

func updateLocalToRemote(s localToRemoteSeam, rt *config.Runtime) (LocalToRemoteResult, error) {
	var result LocalToRemoteResult

	repoPath, _, err := rt.ResolveRepoPath()
	if err != nil {
		return result, err
	}
	if repoPath == "" {
		return result, errors.New("update local-to-remote: no repo configured; clone the macheim repo and set --repo or MACHEIM_REPO first")
	}
	homePath, err := os.UserHomeDir()
	if err != nil {
		return result, fmt.Errorf("locate $HOME: %w", err)
	}

	changes, err := reverseDiff(repoPath, homePath)
	if err != nil {
		return result, err
	}
	renderReverseDrift(changes)
	if len(changes) == 0 {
		return result, nil
	}

	confirmed, err := s.prompt(rt, fmt.Sprintf("Copy these %d changed file(s) back to the repo?", len(changes)))
	if err != nil {
		return result, err
	}
	if !confirmed {
		_, _ = fmt.Fprintln(os.Stderr, "dotfiles: skipped")
		return result, nil
	}

	backupDir := filepath.Join(repoPath, ".macheim-repo-backups", time.Now().UTC().Format("2006-01-02T150405Z"))
	dryRun := rt != nil && rt.DryRun
	verbose := rt != nil && rt.Verbose

	for _, change := range changes {
		if err := copyHomeToRepo(repoPath, homePath, backupDir, change, dryRun, verbose, &result); err != nil {
			return result, err
		}
	}
	return result, nil
}

// reverseChange describes one rel path that drifted in the reverse
// direction (from $HOME's perspective, the repo is out of date). The
// Class field reuses the existing FileClass type, narrowed to the
// values reverseDiff actually emits: Changed (both ends exist, bytes
// differ).
type reverseChange struct {
	RelPath string
	Class   FileClass
}

// reverseDiff walks <repoPath>/dotfiles/ to discover the set of rel
// paths the repo already tracks, then classifies each one against
// $HOME. The semantic is the mirror of Diff: the question is "what
// has changed in $HOME relative to the repo?" — and the only mutation
// we propose is to copy $HOME's version back into the repo when it
// differs.
//
// We do NOT scan $HOME for paths the repo doesn't track. Bringing
// new dotfiles into the repo is a separate workflow (the user must
// decide which $HOME paths to track) and GOALS.md does not specify it.
//
// Missing-in-$HOME paths are not emitted either: GOALS.md's reverse
// direction is "review what's drifted ... and update the REPO files
// to match local state" — the local-vs-repo comparison is byte-level,
// not presence-level. A repo file that no longer exists in $HOME is
// surfaced via the renderer as "missing in $HOME" but is NOT copied
// back (you cannot copy a non-existent file), nor is it removed from
// the repo automatically — that would be destructive without user
// guidance.
func reverseDiff(repoPath, homePath string) ([]reverseChange, error) {
	root := filepath.Join(repoPath, "dotfiles")

	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("dotfiles: stat repo tree: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("dotfiles: %s is not a directory", root)
	}

	var changes []reverseChange
	walkErr := filepath.WalkDir(root, makeReverseWalker(root, homePath, &changes))
	if walkErr != nil {
		return nil, walkErr
	}
	return changes, nil
}

// makeReverseWalker returns the WalkDir callback for reverseDiff,
// extracted to keep reverseDiff under the gocyclo ceiling.
func makeReverseWalker(root, homePath string, changes *[]reverseChange) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && strings.HasSuffix(d.Name(), ".git") {
				return fs.SkipDir
			}
			return nil
		}
		if d.Name() == ".DS_Store" {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		change, ok, classifyErr := classifyReverse(path, filepath.Join(homePath, rel))
		if classifyErr != nil {
			return classifyErr
		}
		if ok {
			*changes = append(*changes, reverseChange{RelPath: rel, Class: change})
		}
		return nil
	}
}

// classifyReverse compares one (repoPath, homePath) pair and reports a
// reverse-direction class. Returns (class, true, nil) when the pair
// produces an actionable change; (_, false, nil) when there is nothing
// to do (identical, or home is missing — see below for the latter).
//
// Cases:
//   - Both exist, bytes/targets equal → (_, false, nil)         no change
//   - Both exist, bytes/targets differ → (Changed, true, nil)    will copy
//   - Home missing → (MissingInRepo, true, nil)                  surfaced only
//   - One is a symlink and the other is not → (Changed, true, nil) will copy
//
// The MissingInRepo result is emitted so the renderer can show the
// user which repo files no longer exist in $HOME, but the copy loop
// skips it (you can't copy a nonexistent file).
func classifyReverse(repoPath, homePath string) (FileClass, bool, error) {
	homeInfo, err := os.Lstat(homePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// File exists in repo but no longer in $HOME. Surface it
			// for visibility; do not act on it automatically.
			return MissingInRepo, true, nil
		}
		return 0, false, fmt.Errorf("dotfiles: lstat %s: %w", homePath, err)
	}
	repoInfo, err := os.Lstat(repoPath)
	if err != nil {
		return 0, false, fmt.Errorf("dotfiles: lstat %s: %w", repoPath, err)
	}

	repoIsLink := repoInfo.Mode()&os.ModeSymlink != 0
	homeIsLink := homeInfo.Mode()&os.ModeSymlink != 0
	if repoIsLink != homeIsLink {
		return Changed, true, nil
	}
	if repoIsLink {
		repoTarget, err := os.Readlink(repoPath)
		if err != nil {
			return 0, false, fmt.Errorf("dotfiles: readlink %s: %w", repoPath, err)
		}
		homeTarget, err := os.Readlink(homePath)
		if err != nil {
			return 0, false, fmt.Errorf("dotfiles: readlink %s: %w", homePath, err)
		}
		if repoTarget == homeTarget {
			return 0, false, nil
		}
		return Changed, true, nil
	}

	repoBytes, err := os.ReadFile(repoPath) //nolint:gosec // path comes from validated repo walk
	if err != nil {
		return 0, false, fmt.Errorf("dotfiles: read %s: %w", repoPath, err)
	}
	homeBytes, err := os.ReadFile(homePath) //nolint:gosec // path comes from validated home dir
	if err != nil {
		return 0, false, fmt.Errorf("dotfiles: read %s: %w", homePath, err)
	}
	if bytes.Equal(repoBytes, homeBytes) {
		return 0, false, nil
	}
	return Changed, true, nil
}

// renderReverseDrift prints a human-readable summary of the diff to
// stderr. Empty input prints "dotfiles: no drift".
func renderReverseDrift(changes []reverseChange) {
	if len(changes) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "dotfiles: no drift")
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, "Dotfiles drift ($HOME -> repo):")
	for _, c := range changes {
		switch c.Class {
		case Changed:
			_, _ = fmt.Fprintf(os.Stderr, "  ~ %s\n", c.RelPath)
		case MissingInRepo:
			// Repo has it, $HOME doesn't. Surface it but we will not act.
			_, _ = fmt.Fprintf(os.Stderr, "  - %s (missing in $HOME; not copied)\n", c.RelPath)
		case Identical, NewInRepo:
			// Not emitted by reverseDiff today, but the switch is
			// exhaustive for the reader.
			continue
		}
	}
}

// copyHomeToRepo handles one rel path's worth of work: back up the
// existing repo file (if present), then copy $HOME's version into
// place. Skips MissingInRepo entries — they were surfaced for
// visibility but cannot be acted on.
//
// Honors dryRun (no fs writes) and verbose (one line per action to
// stderr). Mutates the LocalToRemoteResult through the pointer rather
// than returning slices to keep the loop in updateLocalToRemote thin.
func copyHomeToRepo(repoPath, homePath, backupDir string, change reverseChange, dryRun, verbose bool, result *LocalToRemoteResult) error {
	if change.Class == MissingInRepo {
		return nil
	}
	srcAbs := filepath.Join(homePath, change.RelPath)
	dstAbs := filepath.Join(repoPath, "dotfiles", change.RelPath)
	backupAbs := filepath.Join(backupDir, change.RelPath)

	if err := backupRepoFile(dstAbs, backupAbs, dryRun); err != nil {
		return fmt.Errorf("dotfiles: backup %s: %w", change.RelPath, err)
	}
	result.BackedUp = append(result.BackedUp, change.RelPath)
	if result.BackupDir == "" {
		result.BackupDir = backupDir
	}
	logReverse(verbose, dryRun, "backup: %s -> %s\n", change.RelPath, backupAbs)

	if err := installHomeIntoRepo(srcAbs, dstAbs, dryRun); err != nil {
		return fmt.Errorf("dotfiles: copy %s: %w", change.RelPath, err)
	}
	result.Copied = append(result.Copied, change.RelPath)
	logReverse(verbose, dryRun, "copy: %s -> %s\n", change.RelPath, dstAbs)
	return nil
}

// backupRepoFile copies the existing repo-side file at src to dst
// (under the timestamped backup root), preserving mode. Symlinks are
// backed up as symlinks. Under dryRun it returns nil without touching
// the filesystem.
func backupRepoFile(src, dst string, dryRun bool) error {
	if dryRun {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	return copyRegularFile(src, dst, info.Mode().Perm())
}

// installHomeIntoRepo writes the $HOME-side src into the repo-side dst,
// replacing any pre-existing entry. Symlinks are recreated as symlinks
// at the same target. Under dryRun it returns nil without touching the
// filesystem.
func installHomeIntoRepo(src, dst string, dryRun bool) error {
	if dryRun {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if err := os.Remove(dst); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dst)
	}
	return copyRegularFile(src, dst, info.Mode().Perm())
}

// copyRegularFile copies src to dst with mode. Mirrors apply.go's
// copyRegular — kept as a peer-level helper rather than reused across
// files to keep the forward and reverse direction code independent.
func copyRegularFile(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src) //nolint:gosec // src is a validated path
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, mode); err != nil {
		return err
	}
	return os.Chmod(dst, mode)
}

// logReverse writes one action line to stderr when verbose, prefixed
// `[dry-run] ` under dryRun. Mirrors apply.go's logf.
func logReverse(verbose, dryRun bool, format string, args ...any) {
	if !verbose {
		return
	}
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}
	_, _ = fmt.Fprintf(os.Stderr, prefix+format, args...)
}
