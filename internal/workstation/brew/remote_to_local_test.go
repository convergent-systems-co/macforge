package brew

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
	"github.com/convergent-systems-co/macforge/internal/workstation/embedded"
)

// stubR2LSeam fills panicking defaults so an unstubbed call is loud.
func stubR2LSeam(overrides remoteToLocalSeam) remoteToLocalSeam {
	if overrides.detect == nil {
		overrides.detect = func() (string, bool) { return "/opt/homebrew/bin/brew", true }
	}
	if overrides.run == nil {
		overrides.run = func(*config.Runtime, string, ...string) (string, error) {
			panic("unstubbed run call")
		}
	}
	if overrides.isClean == nil {
		overrides.isClean = func(string) (bool, error) { return true, nil }
	}
	if overrides.pull == nil {
		overrides.pull = func(*config.Runtime, string) error { return nil }
	}
	if overrides.resolveFn == nil {
		overrides.resolveFn = func(string, string) (string, string, error) {
			return "configs/Brewfile", embedded.SourceEmbed, nil
		}
	}
	return overrides
}

func TestUpdateRemoteToLocal_CallsBundleWithRepoBrewfile(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	var (
		ranName string
		ranArgs []string
	)
	s := stubR2LSeam(remoteToLocalSeam{
		run: func(_ *config.Runtime, name string, args ...string) (string, error) {
			ranName = name
			ranArgs = args
			return "", nil
		},
	})
	if err := updateRemoteToLocal(s, &config.Runtime{RepoPath: repo}, false, false); err != nil {
		t.Fatalf("updateRemoteToLocal: %v", err)
	}
	if ranName != "/opt/homebrew/bin/brew" {
		t.Errorf("name = %q, want /opt/homebrew/bin/brew", ranName)
	}
	wantFile := "--file=" + filepath.Join(repo, "Brewfile")
	if len(ranArgs) < 2 || ranArgs[0] != "bundle" || ranArgs[1] != wantFile {
		t.Errorf("args = %v, want [bundle %s ...]", ranArgs, wantFile)
	}
	for _, a := range ranArgs {
		if a == "--cleanup" {
			t.Errorf("--cleanup leaked without --prune; args = %v", ranArgs)
		}
	}
}

func TestUpdateRemoteToLocal_PruneAddsCleanup(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	var ranArgs []string
	s := stubR2LSeam(remoteToLocalSeam{
		run: func(_ *config.Runtime, _ string, args ...string) (string, error) {
			ranArgs = args
			return "", nil
		},
	})
	if err := updateRemoteToLocal(s, &config.Runtime{RepoPath: repo}, true, false); err != nil {
		t.Fatalf("updateRemoteToLocal: %v", err)
	}
	saw := false
	for _, a := range ranArgs {
		if a == "--cleanup" {
			saw = true
			break
		}
	}
	if !saw {
		t.Errorf("--prune did not produce --cleanup; args = %v", ranArgs)
	}
}

func TestUpdateRemoteToLocal_NoPullSkipsPull(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	pullCalled := false
	cleanCalled := false
	s := stubR2LSeam(remoteToLocalSeam{
		run: func(*config.Runtime, string, ...string) (string, error) { return "", nil },
		isClean: func(string) (bool, error) {
			cleanCalled = true
			return true, nil
		},
		pull: func(*config.Runtime, string) error {
			pullCalled = true
			return nil
		},
	})
	if err := updateRemoteToLocal(s, &config.Runtime{RepoPath: repo}, false, true); err != nil {
		t.Fatalf("updateRemoteToLocal: %v", err)
	}
	if pullCalled {
		t.Error("pull called despite --no-pull")
	}
	if cleanCalled {
		t.Error("isClean called despite --no-pull (the dirty-tree guard exists to protect the pull)")
	}
}

func TestUpdateRemoteToLocal_DirtyRepoRefuses(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	var ranCalled bool
	s := stubR2LSeam(remoteToLocalSeam{
		run: func(*config.Runtime, string, ...string) (string, error) {
			ranCalled = true
			return "", nil
		},
		isClean: func(string) (bool, error) { return false, nil },
		pull: func(*config.Runtime, string) error {
			t.Error("pull called despite dirty tree")
			return nil
		},
	})
	err := updateRemoteToLocal(s, &config.Runtime{RepoPath: repo}, false, false)
	if err == nil {
		t.Fatal("expected dirty-tree error, got nil")
	}
	if !strings.Contains(err.Error(), "uncommitted changes") {
		t.Errorf("err = %q, want substring 'uncommitted changes'", err.Error())
	}
	if ranCalled {
		t.Error("brew bundle ran despite dirty-tree refusal")
	}
}

func TestUpdateRemoteToLocal_PullErrorPropagates(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	wantErr := errors.New("network down")
	s := stubR2LSeam(remoteToLocalSeam{
		run:  func(*config.Runtime, string, ...string) (string, error) { return "", nil },
		pull: func(*config.Runtime, string) error { return wantErr },
	})
	err := updateRemoteToLocal(s, &config.Runtime{RepoPath: repo}, false, false)
	if err == nil {
		t.Fatal("expected pull error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want errors.Is wrapping %v", err, wantErr)
	}
}

func TestUpdateRemoteToLocal_BrewNotInstalled(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	s := stubR2LSeam(remoteToLocalSeam{
		detect: func() (string, bool) { return "", false },
	})
	err := updateRemoteToLocal(s, &config.Runtime{RepoPath: repo}, false, false)
	if err == nil {
		t.Fatal("expected error when brew not installed, got nil")
	}
	if !strings.Contains(err.Error(), "brew not installed") {
		t.Errorf("err = %q, want substring 'brew not installed'", err.Error())
	}
}

func TestUpdateRemoteToLocal_NoRepoRefusesEmbedFallback(t *testing.T) {
	// t.Setenv forbids t.Parallel.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("MACHEIM_REPO", "")
	s := stubR2LSeam(remoteToLocalSeam{})
	err := updateRemoteToLocal(s, &config.Runtime{}, false, false)
	if err == nil {
		t.Fatal("expected error in embed-fallback mode, got nil")
	}
	if !strings.Contains(err.Error(), "no repo configured") {
		t.Errorf("err = %q, want substring 'no repo configured'", err.Error())
	}
}

func TestUpdateRemoteToLocal_BundleErrorPropagates(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	wantErr := errors.New("brew bundle failed")
	s := stubR2LSeam(remoteToLocalSeam{
		run: func(*config.Runtime, string, ...string) (string, error) {
			return "", wantErr
		},
	})
	err := updateRemoteToLocal(s, &config.Runtime{RepoPath: repo}, false, false)
	if err == nil {
		t.Fatal("expected bundle error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want errors.Is wrapping %v", err, wantErr)
	}
}
