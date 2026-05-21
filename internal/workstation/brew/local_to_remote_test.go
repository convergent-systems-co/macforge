//go:build !windows

package brew

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// stubL2RSeam fills in panicking defaults for any unset seam fields so
// accidentally calling an unstubbed dependency in a test is loud.
func stubL2RSeam(overrides localToRemoteSeam) localToRemoteSeam {
	if overrides.detect == nil {
		overrides.detect = func() (string, bool) { return "/opt/homebrew/bin/brew", true }
	}
	if overrides.run == nil {
		overrides.run = func(*config.Runtime, string, ...string) (string, error) {
			panic("unstubbed run call")
		}
	}
	if overrides.prompt == nil {
		overrides.prompt = func(*config.Runtime, string) (bool, error) {
			panic("unstubbed prompt call")
		}
	}
	return overrides
}

func TestUpdateLocalToRemote_NoRepoConfigured(t *testing.T) {
	// t.Setenv forbids t.Parallel, so this test runs serially. Sandbox
	// HOME so config-file discovery doesn't latch onto the developer's
	// real ~/.config/macheim/config.yaml during testing.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MACHEIM_REPO", "")
	s := stubL2RSeam(localToRemoteSeam{})
	err := updateLocalToRemote(s, &config.Runtime{})
	if err == nil {
		t.Fatal("expected error in embed-fallback mode, got nil")
	}
	if !strings.Contains(err.Error(), "no repo configured") {
		t.Errorf("err = %q, want substring 'no repo configured'", err.Error())
	}
}

func TestUpdateLocalToRemote_BrewNotInstalled(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	s := stubL2RSeam(localToRemoteSeam{
		detect: func() (string, bool) { return "", false },
	})
	err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo})
	if err == nil {
		t.Fatal("expected error when brew not installed, got nil")
	}
	if !strings.Contains(err.Error(), "brew not installed") {
		t.Errorf("err = %q, want substring 'brew not installed'", err.Error())
	}
}

func TestUpdateLocalToRemote_NoDriftReturnsNil(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	// Both ends have the same single brew entry.
	writeBrewfile(t, filepath.Join(repo, "Brewfile"), `brew "git"`+"\n")
	s := stubL2RSeam(localToRemoteSeam{
		run: func(_ *config.Runtime, _ string, args ...string) (string, error) {
			// Sanity: the dump command is what we expect.
			if !contains(args, "bundle", "dump", "--describe", "--file=-", "--force") {
				t.Errorf("run args = %v, want bundle dump invocation", args)
			}
			return `brew "git"` + "\n", nil
		},
	})
	if err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo}); err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
}

func TestUpdateLocalToRemote_PromptDeclinedSkipsWrite(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	brewfilePath := filepath.Join(repo, "Brewfile")
	writeBrewfile(t, brewfilePath, `brew "git"`+"\n")

	originalContent := readFileContents(t, brewfilePath)

	s := stubL2RSeam(localToRemoteSeam{
		run: func(*config.Runtime, string, ...string) (string, error) {
			// Local state has an extra cask compared to the repo.
			return `brew "git"` + "\n" + `cask "firefox"` + "\n", nil
		},
		prompt: func(*config.Runtime, string) (bool, error) { return false, nil },
	})
	if err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo}); err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if got := readFileContents(t, brewfilePath); got != originalContent {
		t.Errorf("Brewfile changed despite declined prompt:\n got:  %q\n want: %q", got, originalContent)
	}
}

func TestUpdateLocalToRemote_ConfirmRewritesBrewfile(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	brewfilePath := filepath.Join(repo, "Brewfile")
	writeBrewfile(t, brewfilePath, `brew "git"`+"\n")

	s := stubL2RSeam(localToRemoteSeam{
		run: func(*config.Runtime, string, ...string) (string, error) {
			// Local state adds firefox and a new tap.
			return `tap "homebrew/cask"` + "\n" +
				`brew "git"` + "\n" +
				`cask "firefox"` + "\n", nil
		},
		prompt: func(*config.Runtime, string) (bool, error) { return true, nil },
	})
	if err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo}); err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}

	got := readFileContents(t, brewfilePath)
	// Output is grouped by Kind (tap, brew, cask) and sorted by Name
	// within each Kind. The remote git entry was preserved (it appears
	// in Unchanged via the local copy).
	want := `tap "homebrew/cask"` + "\n" +
		`brew "git"` + "\n" +
		`cask "firefox"` + "\n"
	if got != want {
		t.Errorf("rewritten Brewfile:\n got:  %q\n want: %q", got, want)
	}
}

func TestUpdateLocalToRemote_RtYesSkipsPrompt(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	brewfilePath := filepath.Join(repo, "Brewfile")
	writeBrewfile(t, brewfilePath, `brew "git"`+"\n")

	var promptCalled bool
	s := stubL2RSeam(localToRemoteSeam{
		run: func(*config.Runtime, string, ...string) (string, error) {
			return `brew "git"` + "\n" + `cask "firefox"` + "\n", nil
		},
		prompt: func(rt *config.Runtime, _ string) (bool, error) {
			promptCalled = true
			// Mirror the real shell.Prompt contract: rt.Yes makes us return true
			// without reading from stdin.
			if rt != nil && rt.Yes {
				return true, nil
			}
			t.Error("prompt called without rt.Yes; expected short-circuit at the seam layer")
			return false, nil
		},
	})
	if err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo, Yes: true}); err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if !promptCalled {
		t.Error("prompt seam never invoked; rt.Yes path must still go through prompt")
	}
	// Confirm the write occurred.
	got := readFileContents(t, brewfilePath)
	if !strings.Contains(got, `cask "firefox"`) {
		t.Errorf("rewritten Brewfile missing firefox; got %q", got)
	}
}

func TestUpdateLocalToRemote_DryRunSkipsWrite(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	brewfilePath := filepath.Join(repo, "Brewfile")
	original := `brew "git"` + "\n"
	writeBrewfile(t, brewfilePath, original)

	s := stubL2RSeam(localToRemoteSeam{
		run: func(_ *config.Runtime, _ string, _ ...string) (string, error) {
			// Real shell.Run would short-circuit under DryRun returning
			// ("", nil). Our stub returns populated output anyway so the
			// diff is non-empty — this test exercises the writeRepoBrewfile
			// dry-run branch, not shell.Run's.
			return `brew "git"` + "\n" + `cask "firefox"` + "\n", nil
		},
		prompt: func(*config.Runtime, string) (bool, error) { return true, nil },
	})
	if err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo, DryRun: true}); err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	if got := readFileContents(t, brewfilePath); got != original {
		t.Errorf("dry-run mutated Brewfile:\n got:  %q\n want: %q", got, original)
	}
}

func TestUpdateLocalToRemote_MissingRemoteBrewfileTreatedAsEmpty(t *testing.T) {
	t.Parallel()
	repo := t.TempDir() // no Brewfile in repo
	brewfilePath := filepath.Join(repo, "Brewfile")
	s := stubL2RSeam(localToRemoteSeam{
		run: func(*config.Runtime, string, ...string) (string, error) {
			return `brew "git"` + "\n", nil
		},
		prompt: func(*config.Runtime, string) (bool, error) { return true, nil },
	})
	if err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo}); err != nil {
		t.Fatalf("updateLocalToRemote: %v", err)
	}
	got := readFileContents(t, brewfilePath)
	if got != `brew "git"`+"\n" {
		t.Errorf("created Brewfile = %q, want %q", got, `brew "git"`+"\n")
	}
}

func TestUpdateLocalToRemote_RunErrorPropagates(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	wantErr := errors.New("dump failed")
	s := stubL2RSeam(localToRemoteSeam{
		run: func(*config.Runtime, string, ...string) (string, error) {
			return "", wantErr
		},
	})
	err := updateLocalToRemote(s, &config.Runtime{RepoPath: repo})
	if err == nil {
		t.Fatal("expected error from dump, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want errors.Is wrapping %v", err, wantErr)
	}
}

// writeBrewfile is a tiny helper that writes a Brewfile fixture.
func writeBrewfile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil { //nolint:gosec // test fixture
		t.Fatalf("write: %v", err)
	}
}

// readFileContents reads a file and returns the contents as a string.
func readFileContents(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path) //nolint:gosec // test fixture
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(b)
}

// contains reports whether args contains every want token in order.
func contains(args []string, want ...string) bool {
	if len(args) < len(want) {
		return false
	}
	for i := 0; i+len(want) <= len(args); i++ {
		ok := true
		for j := range want {
			if args[i+j] != want[j] {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}