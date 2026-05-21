package dotfiles

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// ApplyResult summarizes one Apply run. The slices are populated in
// repo-walk order. Under --dry-run the result fields populate as if the
// operations had run, but no filesystem mutation occurred.
type ApplyResult struct {
	// Copied is the list of RelPaths that were written (or would have
	// been written under --dry-run).
	Copied []string
	// Skipped is the list of RelPaths that were Identical and required
	// no work.
	Skipped []string
	// BackedUp is the list of RelPaths that had a prior file at $HOME
	// backed up before the new write.
	BackedUp []string
	// BackupDir is the absolute path of the timestamped backup directory.
	// Empty when nothing needed backing up.
	BackupDir string
}

// Apply copies <repoPath>/dotfiles/ over homePath. Any pre-existing file
// at homePath that would be overwritten is first copied to
// homePath/.macheim-backups/<UTC-ISO-8601>/, preserving its relative path
// under the backup root.
//
// Honors rt.DryRun: when true, no filesystem mutation occurs, but the
// returned ApplyResult populates as if the operations had run, so callers
// can preview the change set. The BackupDir field is populated under
// --dry-run too — it names the directory that *would* have been created.
//
// Honors rt.Verbose: when true, each action is logged to os.Stderr
// (prefixed `[dry-run] ` under rt.DryRun).
//
// Honors rt.Yes only insofar as Apply is currently all-or-nothing — there
// is no per-file confirmation in this PR. The future hook for per-file
// prompts lives at the loop body below; when added, rt.Yes will short-
// circuit the prompt the same way internal/shell.Prompt does today.
//
// File mode is preserved (os.Chmod after write). Symlinks are preserved
// as symlinks (Readlink + Symlink); no dereferencing.
func Apply(rt *config.Runtime, repoPath, homePath string) (ApplyResult, error) {
	var result ApplyResult

	entries, err := Diff(repoPath, homePath)
	if err != nil {
		return result, err
	}

	// Pre-compute the backup directory once. We commit to a single
	// timestamped directory per Apply run so all backups from one
	// invocation cluster together (`ls ~/.macheim-backups/` reads as
	// a history of macheim runs, not a history of files).
	backupDir := filepath.Join(homePath, ".macheim-backups", time.Now().UTC().Format("2006-01-02T150405Z"))

	repoRoot := filepath.Join(repoPath, "dotfiles")
	dryRun := rt != nil && rt.DryRun
	verbose := rt != nil && rt.Verbose

	for _, entry := range entries {
		if entry.Class == Identical {
			result.Skipped = append(result.Skipped, entry.RelPath)
			logf(verbose, dryRun, "skip (identical): %s\n", entry.RelPath)
			continue
		}

		// Future hook: per-file confirmation when !rt.Yes would slot
		// in here, calling shell.Prompt(rt, os.Stdin, msg). Today
		// Apply is all-or-nothing; doctor and the eventual
		// `update remote-to-local` command both wrap Apply in their
		// own outer confirmation.

		srcAbs := filepath.Join(repoRoot, entry.RelPath)
		dstAbs := filepath.Join(homePath, entry.RelPath)

		// Backup first, but only when the destination exists (Class
		// == Changed). NewInRepo entries have nothing to back up.
		if entry.Class == Changed {
			backupAbs := filepath.Join(backupDir, entry.RelPath)
			if err := backupFile(dstAbs, backupAbs, dryRun); err != nil {
				return result, fmt.Errorf("dotfiles: backup %s: %w", entry.RelPath, err)
			}
			result.BackedUp = append(result.BackedUp, entry.RelPath)
			if result.BackupDir == "" {
				result.BackupDir = backupDir
			}
			logf(verbose, dryRun, "backup: %s -> %s\n", entry.RelPath, backupAbs)
		}

		if err := installEntry(srcAbs, dstAbs, dryRun); err != nil {
			return result, fmt.Errorf("dotfiles: install %s: %w", entry.RelPath, err)
		}
		result.Copied = append(result.Copied, entry.RelPath)
		logf(verbose, dryRun, "copy: %s -> %s\n", entry.RelPath, dstAbs)
	}

	return result, nil
}

// logf writes one action line to stderr when verbose. The dryRun prefix
// matches internal/shell/run.go's "[dry-run] " convention.
func logf(verbose, dryRun bool, format string, args ...any) {
	if !verbose {
		return
	}
	prefix := ""
	if dryRun {
		prefix = "[dry-run] "
	}
	_, _ = fmt.Fprintf(os.Stderr, prefix+format, args...)
}

// backupFile copies src (an existing file in $HOME) to dst (under the
// timestamped backup root), preserving mode. Under dryRun it returns
// nil without touching the filesystem.
//
// Symlinks are backed up as symlinks (Readlink + Symlink) — the original
// representation is preserved so a restore is byte-for-byte exact.
func backupFile(src, dst string, dryRun bool) error {
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
	return copyRegular(src, dst, info.Mode().Perm())
}

// installEntry writes src into dst, replacing any pre-existing entry.
// Symlinks are recreated as symlinks pointing at the same target.
// Regular files have their mode preserved via os.Chmod after the write.
// Under dryRun it returns nil without touching the filesystem.
func installEntry(src, dst string, dryRun bool) error {
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

	// Remove the destination first so we don't have to special-case
	// "symlink replacing a file" or "file replacing a symlink." Missing
	// is fine; any other error is real and propagates.
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
	return copyRegular(src, dst, info.Mode().Perm())
}

// copyRegular copies src to dst with mode. Errors propagate; partial
// writes leave dst in whatever state os.WriteFile produced (no explicit
// rollback — the next Apply will overwrite or the user can restore from
// the backup directory).
func copyRegular(src, dst string, mode os.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, mode); err != nil {
		return err
	}
	// os.WriteFile honors umask; force the exact source mode.
	return os.Chmod(dst, mode)
}
