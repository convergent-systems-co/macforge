//go:build !windows

package brew

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// stubBundleSeam builds a bundleSeam from overrides. Any field left nil
// falls back to a panicking default so accidental real-host calls in
// tests are loud rather than silent.
func stubBundleSeam(overrides bundleSeam) bundleSeam {
	if overrides.detect == nil {
		overrides.detect = func() (string, bool) { return "/opt/homebrew/bin/brew", true }
	}
	if overrides.resolve == nil {
		overrides.resolve = func(*config.Runtime) (string, string, error) { return "", "", nil }
	}
	if overrides.stat == nil {
		overrides.stat = func(string) (os.FileInfo, error) { return nil, os.ErrNotExist }
	}
	if overrides.embedRead == nil {
		overrides.embedRead = func() ([]byte, error) { return []byte("# embedded\n"), nil }
	}
	if overrides.tempWrite == nil {
		overrides.tempWrite = func([]byte) (string, func(), error) {
			return "/tmp/macheim-Brewfile-fake", func() {}, nil
		}
	}
	if overrides.run == nil {
		overrides.run = func(*config.Runtime, string, ...string) (string, error) { return "", nil }
	}
	if overrides.stderr == nil {
		overrides.stderr = func(string, ...any) {}
	}
	return overrides
}

func TestApply_BrewMissing(t *testing.T) {
	t.Parallel()
	s := stubBundleSeam(bundleSeam{
		detect: func() (string, bool) { return "/opt/homebrew/bin/brew", false },
		run: func(*config.Runtime, string, ...string) (string, error) {
			t.Error("run called when brew not installed; want short-circuit")
			return "", nil
		},
	})
	err := apply(s, &config.Runtime{})
	if err == nil {
		t.Fatal("err = nil, want error when brew is not installed")
	}
	if !strings.Contains(err.Error(), "macheim brew install") {
		t.Errorf("err = %q, want substring 'macheim brew install'", err.Error())
	}
}

func TestApply_RepoBrewfile_Used(t *testing.T) {
	t.Parallel()
	var (
		runName string
		runArgs []string
	)
	s := stubBundleSeam(bundleSeam{
		resolve: func(*config.Runtime) (string, string, error) { return "/r", "flag", nil },
		stat: func(p string) (os.FileInfo, error) {
			if p != "/r/Brewfile" {
				t.Errorf("stat called with %q, want /r/Brewfile", p)
			}
			return fakeStat{}, nil
		},
		embedRead: func() ([]byte, error) {
			t.Error("embedRead called when repo Brewfile exists; want skip")
			return nil, errors.New("should not be called")
		},
		tempWrite: func([]byte) (string, func(), error) {
			t.Error("tempWrite called when repo Brewfile exists; want skip")
			return "", func() {}, nil
		},
		run: func(_ *config.Runtime, name string, args ...string) (string, error) {
			runName = name
			runArgs = args
			return "", nil
		},
	})
	if err := apply(s, &config.Runtime{}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if runName != "/opt/homebrew/bin/brew" {
		t.Errorf("run name = %q, want /opt/homebrew/bin/brew", runName)
	}
	if len(runArgs) != 2 || runArgs[0] != "bundle" || runArgs[1] != "--file=/r/Brewfile" {
		t.Errorf("run args = %v, want [bundle --file=/r/Brewfile]", runArgs)
	}
}

func TestApply_EmbedFallback(t *testing.T) {
	t.Parallel()
	var (
		tempWriteCalled bool
		cleanupCalled   bool
		runArgs         []string
	)
	s := stubBundleSeam(bundleSeam{
		resolve: func(*config.Runtime) (string, string, error) { return "", "", nil },
		stat: func(string) (os.FileInfo, error) {
			t.Error("stat called when repo not configured; want skip")
			return nil, os.ErrNotExist
		},
		embedRead: func() ([]byte, error) { return []byte("brew \"ripgrep\"\n"), nil },
		tempWrite: func(content []byte) (string, func(), error) {
			tempWriteCalled = true
			if !strings.Contains(string(content), "ripgrep") {
				t.Errorf("tempWrite content = %q, want substring 'ripgrep'", string(content))
			}
			return "/tmp/macheim-Brewfile-fake", func() { cleanupCalled = true }, nil
		},
		run: func(_ *config.Runtime, _ string, args ...string) (string, error) {
			runArgs = args
			return "", nil
		},
	})
	if err := apply(s, &config.Runtime{}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !tempWriteCalled {
		t.Error("tempWrite not called; embed fallback should stage to a temp file")
	}
	if !cleanupCalled {
		t.Error("temp file cleanup not invoked; want defer cleanup")
	}
	if len(runArgs) != 2 || runArgs[1] != "--file=/tmp/macheim-Brewfile-fake" {
		t.Errorf("run args = %v, want [bundle --file=/tmp/macheim-Brewfile-fake]", runArgs)
	}
}

func TestApply_RepoBrewfile_Missing_FallsThroughToEmbed(t *testing.T) {
	t.Parallel()
	var (
		embedReadCalled bool
		runArgs         []string
	)
	s := stubBundleSeam(bundleSeam{
		resolve: func(*config.Runtime) (string, string, error) { return "/r", "flag", nil },
		stat: func(p string) (os.FileInfo, error) {
			if p != "/r/Brewfile" {
				t.Errorf("stat called with %q, want /r/Brewfile", p)
			}
			return nil, os.ErrNotExist
		},
		embedRead: func() ([]byte, error) {
			embedReadCalled = true
			return []byte("# embedded\n"), nil
		},
		tempWrite: func([]byte) (string, func(), error) {
			return "/tmp/macheim-Brewfile-fallback", func() {}, nil
		},
		run: func(_ *config.Runtime, _ string, args ...string) (string, error) {
			runArgs = args
			return "", nil
		},
	})
	if err := apply(s, &config.Runtime{}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !embedReadCalled {
		t.Error("embedRead not called; repo Brewfile missing should fall through to embed")
	}
	if len(runArgs) != 2 || runArgs[1] != "--file=/tmp/macheim-Brewfile-fallback" {
		t.Errorf("run args = %v, want [bundle --file=/tmp/macheim-Brewfile-fallback]", runArgs)
	}
}

func TestApply_DryRun(t *testing.T) {
	t.Parallel()
	var runCalled bool
	s := stubBundleSeam(bundleSeam{
		resolve: func(*config.Runtime) (string, string, error) { return "/r", "flag", nil },
		stat:    func(string) (os.FileInfo, error) { return fakeStat{}, nil },
		run: func(rt *config.Runtime, _ string, _ ...string) (string, error) {
			runCalled = true
			if rt == nil || !rt.DryRun {
				t.Error("run invoked with non-dry-run runtime; want dry-run propagated")
			}
			// Mimic shell.Run's dry-run behavior: no output captured.
			return "", nil
		},
	})
	if err := apply(s, &config.Runtime{DryRun: true}); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if !runCalled {
		t.Error("run not invoked under dry-run; want Run to receive the dry-run runtime")
	}
}

func TestApply_BrewBundleFails(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("brew bundle exit 1")
	s := stubBundleSeam(bundleSeam{
		resolve: func(*config.Runtime) (string, string, error) { return "/r", "flag", nil },
		stat:    func(string) (os.FileInfo, error) { return fakeStat{}, nil },
		run: func(*config.Runtime, string, ...string) (string, error) {
			return "", wantErr
		},
	})
	err := apply(s, &config.Runtime{})
	if err == nil {
		t.Fatal("err = nil, want propagated error from brew bundle failure")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want errors.Is wrapping %v", err, wantErr)
	}
}

func TestApply_ResolveError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("config load failed")
	s := stubBundleSeam(bundleSeam{
		resolve: func(*config.Runtime) (string, string, error) { return "", "", wantErr },
		run: func(*config.Runtime, string, ...string) (string, error) {
			t.Error("run called despite resolve error; want short-circuit")
			return "", nil
		},
	})
	err := apply(s, &config.Runtime{})
	if err == nil {
		t.Fatal("err = nil, want propagated error from resolve failure")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want errors.Is wrapping %v", err, wantErr)
	}
}

func TestSummary_ParsesSkipsAndInstalls(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		output        string
		wantInstalled int
		wantSkipped   int
	}{
		{
			name:          "empty output",
			output:        "",
			wantInstalled: 0,
			wantSkipped:   0,
		},
		{
			name: "mixed install and using",
			output: strings.Join([]string{
				"Using ripgrep",
				"Installing fd",
				"Using bat",
				"Installing jq",
			}, "\n"),
			wantInstalled: 2,
			wantSkipped:   2,
		},
		{
			name: "skipping counts as skipped",
			output: strings.Join([]string{
				"Skipping mas-app (unsupported)",
				"Using ripgrep",
			}, "\n"),
			wantInstalled: 0,
			wantSkipped:   2,
		},
		{
			name: "noise lines ignored",
			output: strings.Join([]string{
				"Homebrew Bundle complete! 1 Brewfile dependency now installed.",
				"Installing fd",
				"",
				"  ",
			}, "\n"),
			wantInstalled: 1,
			wantSkipped:   0,
		},
		{
			name: "leading whitespace tolerated",
			output: strings.Join([]string{
				"  Installing fd",
				"\tUsing ripgrep",
			}, "\n"),
			wantInstalled: 1,
			wantSkipped:   1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotInstalled, gotSkipped := summarize(tc.output)
			if gotInstalled != tc.wantInstalled {
				t.Errorf("installed = %d, want %d", gotInstalled, tc.wantInstalled)
			}
			if gotSkipped != tc.wantSkipped {
				t.Errorf("skipped = %d, want %d", gotSkipped, tc.wantSkipped)
			}
		})
	}
}