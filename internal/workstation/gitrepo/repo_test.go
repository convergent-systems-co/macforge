//go:build !windows

package gitrepo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// newRepo creates a tempdir, runs `git init`, configures a local identity,
// and lands a single "init" commit so LastCommit/IsClean have something to
// observe. The repo has no remote, so Pull will fail unless DryRun is set.
func newRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, cmd := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "commit.gpgsign", "false"},
	} {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(cmd, " "), err, string(out))
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("init\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, cmd := range [][]string{
		{"git", "add", "README.md"},
		{"git", "commit", "-m", "init"},
	} {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = dir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", strings.Join(cmd, " "), err, string(out))
		}
	}
	return dir
}

func TestLastCommit(t *testing.T) {
	dir := newRepo(t)
	sha, subject, isoDate, err := LastCommit(dir)
	if err != nil {
		t.Fatalf("LastCommit: %v", err)
	}
	if sha == "" {
		t.Errorf("sha is empty")
	}
	if subject != "init" {
		t.Errorf("subject = %q, want %q", subject, "init")
	}
	if _, err := time.Parse(time.RFC3339, isoDate); err != nil {
		t.Errorf("isoDate %q is not RFC3339: %v", isoDate, err)
	}
}

func TestIsClean_Yes(t *testing.T) {
	dir := newRepo(t)
	clean, err := IsClean(dir)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if !clean {
		t.Errorf("clean = false, want true on a fresh repo")
	}
}

func TestIsClean_No(t *testing.T) {
	dir := newRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("dirty\n"), 0644); err != nil {
		t.Fatal(err)
	}
	clean, err := IsClean(dir)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if clean {
		t.Errorf("clean = true, want false after modifying tracked file")
	}
}

func TestPull_DryRunSkipsExec(t *testing.T) {
	dir := newRepo(t)
	rt := &config.Runtime{DryRun: true}
	// The repo has no upstream; a real `git pull` would fail. The fact that
	// Pull returns nil here is the evidence that dry-run short-circuited
	// before exec.
	if err := Pull(rt, dir); err != nil {
		t.Fatalf("Pull dry-run: %v", err)
	}
}

func TestPull_NoUpstreamErrors(t *testing.T) {
	dir := newRepo(t)
	rt := &config.Runtime{}
	err := Pull(rt, dir)
	if err == nil {
		t.Fatal("Pull on repo without upstream: want error, got nil")
	}
	if !strings.Contains(err.Error(), "git pull --ff-only failed") {
		t.Errorf("error %q lacks the wrapped prefix", err.Error())
	}
}