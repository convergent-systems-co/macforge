// Package dotfiles implements the diff/apply machinery that compares a
// repo's `dotfiles/` tree against the user's $HOME and (under apply)
// copies the repo tree into $HOME with backups.
//
// The repo layout is conventional: every path beneath `<repoPath>/dotfiles/`
// maps one-to-one to the same relative path under $HOME. For example,
// `<repoPath>/dotfiles/.zshrc` corresponds to `$HOME/.zshrc`, and
// `<repoPath>/dotfiles/.config/macheim/config.yaml` corresponds to
// `$HOME/.config/macheim/config.yaml`.
//
// Symlinks in the repo are preserved as symlinks in $HOME — Diff and
// Apply both use os.Lstat / os.Readlink rather than dereferencing.
package dotfiles

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileClass classifies one path in the diff between a repo dotfiles tree
// and $HOME.
type FileClass int

const (
	// Identical means the path exists in both trees with byte-equal
	// content (or, for symlinks, identical targets).
	Identical FileClass = iota
	// Changed means the path exists in both trees but the contents
	// differ. Apply would overwrite $HOME's copy and back the old
	// version up first.
	Changed
	// NewInRepo means the path exists in the repo tree but is missing
	// from $HOME. Apply would create it; nothing to back up.
	NewInRepo
	// MissingInRepo describes the reverse direction — present in $HOME's
	// dotfile set but no longer in the repo. The forward Diff (this file)
	// never emits it; it exists in the type for symmetry with the
	// future reverse-direction diff that powers `update local-to-remote`
	// for dotfiles (sub-issue #18).
	MissingInRepo
)

// String renders a FileClass for log lines and tests. The strings are
// stable; tests and renderers may match on them.
func (c FileClass) String() string {
	switch c {
	case Identical:
		return "identical"
	case Changed:
		return "changed"
	case NewInRepo:
		return "new-in-repo"
	case MissingInRepo:
		return "missing-in-repo"
	default:
		return fmt.Sprintf("FileClass(%d)", int(c))
	}
}

// DiffEntry is one row of the diff.
type DiffEntry struct {
	// RelPath is the path relative to the repo's dotfiles/ directory,
	// using forward slashes on every platform (filepath.WalkDir already
	// returns OS-native separators; on darwin the two are the same).
	RelPath string
	// Class is the classification of this entry against $HOME.
	Class FileClass
}

// Diff compares <repoPath>/dotfiles/ against the corresponding paths
// under homePath. It returns the list of entries in repo-walk order;
// only Identical, Changed, and NewInRepo can surface from this
// direction. (MissingInRepo is reserved for the reverse direction —
// sub-issue #18.)
//
// `.DS_Store` files and any directory whose name ends in `.git` are
// skipped entirely. Directories themselves are not emitted — only files
// and symlinks.
//
// Returns an error if the repo's dotfiles tree is missing or unreadable.
// A repo with no `dotfiles/` directory at all is an empty input: the
// function returns (nil, nil).
func Diff(repoPath, homePath string) ([]DiffEntry, error) {
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

	var entries []DiffEntry
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		// Skip junk + nested git checkouts at any depth.
		if d.IsDir() {
			if path == root {
				return nil
			}
			if strings.HasSuffix(name, ".git") {
				return fs.SkipDir
			}
			return nil
		}
		if name == ".DS_Store" {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}

		class, classifyErr := classify(path, filepath.Join(homePath, rel))
		if classifyErr != nil {
			return classifyErr
		}
		entries = append(entries, DiffEntry{RelPath: rel, Class: class})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}
	return entries, nil
}

// classify produces the FileClass for a single (repo, home) pair. Both
// paths are absolute. repoPath is guaranteed to exist (the walker
// produced it); homePath may not.
//
// Symlinks compare by target, not by what the target resolves to —
// this preserves the "store the link, not the link's contents" property
// that Apply also respects.
func classify(repoPath, homePath string) (FileClass, error) {
	homeInfo, err := os.Lstat(homePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return NewInRepo, nil
		}
		return 0, fmt.Errorf("dotfiles: lstat %s: %w", homePath, err)
	}

	repoInfo, err := os.Lstat(repoPath)
	if err != nil {
		return 0, fmt.Errorf("dotfiles: lstat %s: %w", repoPath, err)
	}

	repoIsLink := repoInfo.Mode()&os.ModeSymlink != 0
	homeIsLink := homeInfo.Mode()&os.ModeSymlink != 0

	// A symlink and a regular file are never identical, even if the
	// link's target happens to share the file's content. The user has
	// chosen one representation in the repo; Apply will install that
	// representation in $HOME.
	if repoIsLink != homeIsLink {
		return Changed, nil
	}

	if repoIsLink {
		repoTarget, err := os.Readlink(repoPath)
		if err != nil {
			return 0, fmt.Errorf("dotfiles: readlink %s: %w", repoPath, err)
		}
		homeTarget, err := os.Readlink(homePath)
		if err != nil {
			return 0, fmt.Errorf("dotfiles: readlink %s: %w", homePath, err)
		}
		if repoTarget == homeTarget {
			return Identical, nil
		}
		return Changed, nil
	}

	// Regular file: byte-compare. We rely on os.ReadFile rather than
	// streaming because dotfiles are typically small (kilobytes); if
	// someone checks a 100 MB file into their dotfiles repo, the
	// failure mode is "uses a bit of RAM," not "wrong answer."
	repoBytes, err := os.ReadFile(repoPath)
	if err != nil {
		return 0, fmt.Errorf("dotfiles: read %s: %w", repoPath, err)
	}
	homeBytes, err := os.ReadFile(homePath)
	if err != nil {
		return 0, fmt.Errorf("dotfiles: read %s: %w", homePath, err)
	}
	if bytes.Equal(repoBytes, homeBytes) {
		return Identical, nil
	}
	return Changed, nil
}
