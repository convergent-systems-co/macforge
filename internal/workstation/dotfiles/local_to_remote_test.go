package dotfiles

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// stubL2RSeam fills in a panicking prompt default so an unstubbed
// prompt call in a test is loud.
func stubL2RSeam(overrides localToRemoteSeam) localToRemoteSeam {
	if overrides.prompt == nil {
		overrides.prompt = func(*config.Runtime, string) (bool, error) {
			panic("unstubbed prompt call")
		}
	}
	return overrides
}

func TestUpdateLocalToRemote_NoRepoConfigured(t *testing.T) {
	// t.Setenv forbids t.Parallel.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MACHEIM_REPO", "")
	s := stubL2RSeam(localToRemoteSeam{})
	_, err := updateLocalToRemote(s, &config.Runtime{})
	if err == nil {
		t.Fatal("expected error in embed-fallback mode, got nil")
	}
	if !strings.Contains(err.Error(), "no repo configured") {
		t.Errorf("err = %q, want substring 'no repo configured'", err.Error())
	}
}

// TestUpdateLocalToRemote_* tests below use t.Setenv for HOME, so each
// is non-parallel. We accept the serial penalty (these are small tests)
// in exchange for not refactoring updateLocalToRemote to accept a homeDir.

func TestUpdateLocalToRemote_NoDrift(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "x\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "x\n", 0o644)

	s := stubL2RSeam(localToRemoteSeam{})
	result, err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo})
	if err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if len(result.Copied) != 0 {
		t.Errorf("Copied = %v, want empty", result.Copied)
	}
	if len(result.BackedUp) != 0 {
		t.Errorf("BackedUp = %v, want empty", result.BackedUp)
	}
}

func TestUpdateLocalToRemote_CopiesChangedFile(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "old\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "new\n", 0o644)

	s := stubL2RSeam(localToRemoteSeam{
		prompt: func(*config.Runtime, string) (bool, error) { return true, nil },
	})
	result, err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo})
	if err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if len(result.Copied) != 1 || result.Copied[0] != ".zshrc" {
		t.Errorf("Copied = %v, want [.zshrc]", result.Copied)
	}
	if len(result.BackedUp) != 1 || result.BackedUp[0] != ".zshrc" {
		t.Errorf("BackedUp = %v, want [.zshrc]", result.BackedUp)
	}
	if result.BackupDir == "" {
		t.Fatal("expected BackupDir to be populated")
	}
	if !strings.HasPrefix(result.BackupDir, filepath.Join(repo, ".macheim-repo-backups")) {
		t.Errorf("BackupDir = %q; expected under %s", result.BackupDir, repo)
	}

	// The repo's .zshrc now matches $HOME.
	got, err := os.ReadFile(filepath.Join(repo, "dotfiles", ".zshrc"))
	if err != nil {
		t.Fatalf("read post-copy repo .zshrc: %v", err)
	}
	if string(got) != "new\n" {
		t.Errorf("post-copy repo .zshrc = %q, want %q", string(got), "new\n")
	}
	// The backup retains the old content.
	backupPath := filepath.Join(result.BackupDir, ".zshrc")
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backupContent) != "old\n" {
		t.Errorf("backup = %q, want %q", string(backupContent), "old\n")
	}
}

func TestUpdateLocalToRemote_PromptDeclinedSkips(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "old\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "new\n", 0o644)

	s := stubL2RSeam(localToRemoteSeam{
		prompt: func(*config.Runtime, string) (bool, error) { return false, nil },
	})
	result, err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo})
	if err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if len(result.Copied) != 0 {
		t.Errorf("Copied = %v, want empty (prompt declined)", result.Copied)
	}
	// Repo content should be unchanged.
	got, _ := os.ReadFile(filepath.Join(repo, "dotfiles", ".zshrc"))
	if string(got) != "old\n" {
		t.Errorf("repo file mutated despite decline: %q", string(got))
	}
}

func TestUpdateLocalToRemote_DryRunDoesNotWrite(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "old\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "new\n", 0o644)

	s := stubL2RSeam(localToRemoteSeam{
		prompt: func(*config.Runtime, string) (bool, error) { return true, nil },
	})
	result, err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo, DryRun: true})
	if err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	// Result still populates so the user can preview.
	if len(result.Copied) != 1 {
		t.Errorf("Copied = %v, want 1 entry under dry-run", result.Copied)
	}
	// Repo and backup dir untouched on disk.
	got, _ := os.ReadFile(filepath.Join(repo, "dotfiles", ".zshrc"))
	if string(got) != "old\n" {
		t.Errorf("dry-run mutated repo file: %q", string(got))
	}
	if _, err := os.Stat(filepath.Join(repo, ".macheim-repo-backups")); !os.IsNotExist(err) {
		t.Errorf("dry-run created backup directory: %v", err)
	}
}

func TestUpdateLocalToRemote_MissingInHomeSurfacedNotCopied(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(repo, "dotfiles", ".gone"), "x\n", 0o644)
	// $HOME deliberately empty.

	s := stubL2RSeam(localToRemoteSeam{
		prompt: func(*config.Runtime, string) (bool, error) { return true, nil },
	})
	result, err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo})
	if err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if len(result.Copied) != 0 {
		t.Errorf("Copied = %v, want empty (file missing in $HOME)", result.Copied)
	}
	// File still in repo, untouched.
	got, _ := os.ReadFile(filepath.Join(repo, "dotfiles", ".gone"))
	if string(got) != "x\n" {
		t.Errorf("repo file mutated: %q", string(got))
	}
}

func TestUpdateLocalToRemote_RtYesSkipsPrompt(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeFile(t, filepath.Join(repo, "dotfiles", ".zshrc"), "old\n", 0o644)
	writeFile(t, filepath.Join(home, ".zshrc"), "new\n", 0o644)

	var promptCalled bool
	s := stubL2RSeam(localToRemoteSeam{
		prompt: func(rt *config.Runtime, _ string) (bool, error) {
			promptCalled = true
			// Mirror the real shell.Prompt: rt.Yes returns true.
			if rt != nil && rt.Yes {
				return true, nil
			}
			t.Error("prompt asked without rt.Yes")
			return false, nil
		},
	})
	result, err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo, Yes: true})
	if err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if !promptCalled {
		t.Error("prompt seam not invoked")
	}
	if len(result.Copied) != 1 {
		t.Errorf("Copied = %v, want 1 entry", result.Copied)
	}
}

func TestUpdateLocalToRemote_PreservesSymlinkTargetChange(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	makeSymlink(t, "/old/target", filepath.Join(repo, "dotfiles", ".link"))
	makeSymlink(t, "/new/target", filepath.Join(home, ".link"))

	s := stubL2RSeam(localToRemoteSeam{
		prompt: func(*config.Runtime, string) (bool, error) { return true, nil },
	})
	result, err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo})
	if err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if len(result.Copied) != 1 || result.Copied[0] != ".link" {
		t.Errorf("Copied = %v, want [.link]", result.Copied)
	}
	target, err := os.Readlink(filepath.Join(repo, "dotfiles", ".link"))
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != "/new/target" {
		t.Errorf("new repo link target = %q, want /new/target", target)
	}
}
