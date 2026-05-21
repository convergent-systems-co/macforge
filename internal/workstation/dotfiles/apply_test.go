package dotfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/polliard/macheim/internal/config"
)

// readFile is a tiny helper to keep test assertions readable.
func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// containsStr fails the test if want is not in haystack.
func containsStr(t *testing.T, haystack []string, want string) {
	t.Helper()
	for _, s := range haystack {
		if s == want {
			return
		}
	}
	t.Fatalf("expected %q in %v", want, haystack)
}

func TestApply_CopiesNewFile(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "export PATH=/x\n", 0o640)

	rt := &config.Runtime{}
	result, err := Apply(rt, repo, home)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	containsStr(t, result.Copied, ".zshrc")
	if len(result.BackedUp) != 0 {
		t.Fatalf("expected no backups, got %v", result.BackedUp)
	}

	got := readFile(t, filepath.Join(home, ".zshrc"))
	if got != "export PATH=/x\n" {
		t.Fatalf("content = %q", got)
	}
	// Mode preservation (umask-independent).
	info, err := os.Stat(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("mode = %o, want 0640", info.Mode().Perm())
	}
}

func TestApply_BacksUpChangedFile(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "new\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "old\n", 0o644)

	rt := &config.Runtime{}
	result, err := Apply(rt, repo, home)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	containsStr(t, result.Copied, ".zshrc")
	containsStr(t, result.BackedUp, ".zshrc")
	if result.BackupDir == "" {
		t.Fatal("expected BackupDir to be populated")
	}

	if got := readFile(t, filepath.Join(home, ".zshrc")); got != "new\n" {
		t.Fatalf("post-apply $HOME content = %q", got)
	}
	backupPath := filepath.Join(result.BackupDir, ".zshrc")
	if got := readFile(t, backupPath); got != "old\n" {
		t.Fatalf("backup content = %q", got)
	}
	if !strings.HasPrefix(result.BackupDir, filepath.Join(home, ".macheim-backups")) {
		t.Fatalf("BackupDir = %q; expected under %s", result.BackupDir, home)
	}
}

func TestApply_SkipsIdentical(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "same\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "same\n", 0o644)

	// Snapshot mtime so we can verify Apply did not rewrite.
	infoBefore, err := os.Stat(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	rt := &config.Runtime{}
	result, err := Apply(rt, repo, home)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	containsStr(t, result.Skipped, ".zshrc")
	if len(result.Copied) != 0 {
		t.Fatalf("expected nothing copied, got %v", result.Copied)
	}
	if len(result.BackedUp) != 0 {
		t.Fatalf("expected nothing backed up, got %v", result.BackedUp)
	}
	infoAfter, err := os.Stat(filepath.Join(home, ".zshrc"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if !infoAfter.ModTime().Equal(infoBefore.ModTime()) {
		t.Fatalf("identical file was rewritten (mtime changed)")
	}
	// No backup directory should have been created.
	if _, err := os.Stat(filepath.Join(home, ".macheim-backups")); !os.IsNotExist(err) {
		t.Fatalf(".macheim-backups exists when nothing was backed up: %v", err)
	}
}

func TestApply_PreservesSymlinks(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	makeSymlink(t, "/foo/bar", filepath.Join(repo, "dotfiles", ".link"))

	rt := &config.Runtime{}
	result, err := Apply(rt, repo, home)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	containsStr(t, result.Copied, ".link")

	dst := filepath.Join(home, ".link")
	info, err := os.Lstat(dst)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected symlink at %s, got mode %v", dst, info.Mode())
	}
	target, err := os.Readlink(dst)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != "/foo/bar" {
		t.Fatalf("target = %q, want /foo/bar", target)
	}
}

func TestApply_DryRun(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "new\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "old\n", 0o644)
	writeFile(t, filepath.Join(repo, "dotfiles", ".vimrc"), "set nocompatible\n", 0o644)
	// .vimrc is new in repo, missing in $HOME.

	rt := &config.Runtime{DryRun: true}
	result, err := Apply(rt, repo, home)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	containsStr(t, result.Copied, ".zshrc")
	containsStr(t, result.Copied, ".vimrc")
	containsStr(t, result.BackedUp, ".zshrc")
	if result.BackupDir == "" {
		t.Fatal("expected BackupDir to be populated under --dry-run")
	}

	// Nothing on disk should have changed.
	if got := readFile(t, filepath.Join(home, ".zshrc")); got != "old\n" {
		t.Fatalf("dry-run mutated .zshrc: %q", got)
	}
	if _, err := os.Stat(filepath.Join(home, ".vimrc")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created .vimrc: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".macheim-backups")); !os.IsNotExist(err) {
		t.Fatalf("dry-run created backup directory: %v", err)
	}
}
