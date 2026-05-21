package dotfiles

import (
	"os"
	"path/filepath"
	"testing"
)

// writeFile writes content to path with mode, creating parents as needed.
// Test helper; panics on failure because a test that can't set up its own
// fixture has no recovery path.
func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// makeSymlink creates a symlink at path pointing to target.
func makeSymlink(t *testing.T, target, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("symlink %s -> %s: %v", path, target, err)
	}
}

// findEntry returns the DiffEntry with the given RelPath, or fails the
// test. We hide the slice search so tests can read as assertions about
// behavior, not iteration.
func findEntry(t *testing.T, entries []DiffEntry, rel string) DiffEntry {
	t.Helper()
	for _, e := range entries {
		if e.RelPath == rel {
			return e
		}
	}
	t.Fatalf("expected entry for %q in %+v", rel, entries)
	return DiffEntry{}
}

func TestDiff_Identical(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "export A=1\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "export A=1\n", 0o644)

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	got := findEntry(t, entries, ".zshrc")
	if got.Class != Identical {
		t.Fatalf("class = %v, want Identical", got.Class)
	}
}

func TestDiff_Changed(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "new\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "old\n", 0o644)

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	got := findEntry(t, entries, ".zshrc")
	if got.Class != Changed {
		t.Fatalf("class = %v, want Changed", got.Class)
	}
}

func TestDiff_NewInRepo(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "x\n", 0o644)
	// $HOME deliberately empty.

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	got := findEntry(t, entries, ".zshrc")
	if got.Class != NewInRepo {
		t.Fatalf("class = %v, want NewInRepo", got.Class)
	}
}

func TestDiff_Symlink_Identical(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	makeSymlink(t, "/foo/bar", filepath.Join(repo, "dotfiles", ".link"))
	makeSymlink(t, "/foo/bar", filepath.Join(home, ".link"))

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	got := findEntry(t, entries, ".link")
	if got.Class != Identical {
		t.Fatalf("class = %v, want Identical", got.Class)
	}
}

func TestDiff_Symlink_Differ(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	makeSymlink(t, "/foo/new", filepath.Join(repo, "dotfiles", ".link"))
	makeSymlink(t, "/foo/old", filepath.Join(home, ".link"))

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	got := findEntry(t, entries, ".link")
	if got.Class != Changed {
		t.Fatalf("class = %v, want Changed", got.Class)
	}
}

func TestDiff_NestedFiles(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".config", "foo", "bar.yaml"), "k: v\n", 0o644)

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	rel := filepath.Join(".config", "foo", "bar.yaml")
	got := findEntry(t, entries, rel)
	if got.Class != NewInRepo {
		t.Fatalf("class = %v, want NewInRepo", got.Class)
	}
}

func TestDiff_SkipsDotDS_Store(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".DS_Store"), "junk", 0o644)
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "x\n", 0o644)

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	for _, e := range entries {
		if e.RelPath == ".DS_Store" {
			t.Fatalf(".DS_Store leaked into diff: %+v", entries)
		}
	}
	// Sanity: .zshrc still shows up so the walker did run.
	_ = findEntry(t, entries, ".zshrc")
}

func TestDiff_SkipsGitDirs(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	// A nested .git directory should be pruned entirely.
	writeFile(t, filepath.Join(repo, "dotfiles", ".config", "repo", ".git", "HEAD"), "ref: x\n", 0o644)
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "x\n", 0o644)

	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	for _, e := range entries {
		if filepath.Base(filepath.Dir(e.RelPath)) == ".git" {
			t.Fatalf(".git child leaked: %+v", entries)
		}
	}
	_ = findEntry(t, entries, ".zshrc")
}

func TestDiff_MissingRepoTree(t *testing.T) {
	// A repo with no dotfiles/ directory must not error — it's the
	// embed-fallback case for an early bootstrap.
	repo := t.TempDir()
	home := t.TempDir()
	entries, err := Diff(repo, home)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty entries, got %+v", entries)
	}
}
