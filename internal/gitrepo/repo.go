// Package gitrepo wraps os/exec calls to the system `git` binary. Per
// GOALS.md, we deliberately avoid go-git: the operations are small and
// reading from the user's git CLI matches the user's mental model.
package gitrepo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/polliard/macheim/internal/config"
)

// Pull runs `git -C <repoPath> pull --ff-only`. Honors rt.DryRun: in dry-run
// mode it prints the would-be command to stderr and returns nil without
// invoking git. Returns an error when the pull fails (typically a
// non-fast-forward, missing upstream, or network issue); the underlying
// git output is wrapped into the error for the user.
func Pull(rt *config.Runtime, repoPath string) error {
	if rt != nil && rt.DryRun {
		_, _ = fmt.Fprintf(os.Stderr, "[dry-run] git -C %s pull --ff-only\n", repoPath)
		return nil
	}
	cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull --ff-only failed in %s: %w (%s)", repoPath, err, strings.TrimSpace(string(output)))
	}
	return nil
}

// LastCommit returns the SHA, subject, and ISO-8601 commit date of the most
// recent commit on HEAD in repoPath. The format is produced by
// `git log -1 --format=%H%n%s%n%cI`, which separates the three fields by
// newline so a subject containing tabs or spaces survives intact.
func LastCommit(repoPath string) (sha, subject, isoDate string, err error) {
	cmd := exec.Command("git", "-C", repoPath, "log", "-1", "--format=%H%n%s%n%cI")
	out, runErr := cmd.Output()
	if runErr != nil {
		return "", "", "", fmt.Errorf("git log -1 failed in %s: %w", repoPath, runErr)
	}
	// Trim the trailing newline git appends, then split into the three fields.
	parts := strings.SplitN(strings.TrimRight(string(out), "\n"), "\n", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("git log -1 produced %d field(s), want 3 (output: %q)", len(parts), string(out))
	}
	sha, subject, isoDate = parts[0], parts[1], parts[2]
	if sha == "" || subject == "" || isoDate == "" {
		return "", "", "", fmt.Errorf("git log -1 returned empty field(s): sha=%q subject=%q isoDate=%q", sha, subject, isoDate)
	}
	return sha, subject, isoDate, nil
}

// IsClean reports whether the working tree at repoPath has no pending
// changes. It is true iff `git status --porcelain` produces no output.
func IsClean(repoPath string) (bool, error) {
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status --porcelain failed in %s: %w", repoPath, err)
	}
	return strings.TrimSpace(string(out)) == "", nil
}
