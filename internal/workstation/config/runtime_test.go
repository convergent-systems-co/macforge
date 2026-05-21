package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntime_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		verbose bool
		quiet   bool
		wantErr string
	}{
		{name: "both off", verbose: false, quiet: false, wantErr: ""},
		{name: "only verbose", verbose: true, quiet: false, wantErr: ""},
		{name: "only quiet", verbose: false, quiet: true, wantErr: ""},
		{name: "both on", verbose: true, quiet: true, wantErr: "mutually exclusive"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rt := &Runtime{Verbose: tc.verbose, Quiet: tc.quiet}
			err := rt.Validate()
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("want nil error, got %v", err)
			case tc.wantErr != "" && err == nil:
				t.Fatalf("want error containing %q, got nil", tc.wantErr)
			case tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr):
				t.Fatalf("want error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

// resolveCase describes one row of the discovery-chain table. The fixture
// callback runs after HOME is set to a fresh tempdir, so it can mkdir
// convention paths or write config.yaml under that HOME freely.
type resolveCase struct {
	name         string
	repoFlag     string
	envValue     string
	fixture      func(t *testing.T, home string)
	wantPath     string // empty + wantSource empty => embed-fallback
	wantPathFunc bool   // when true, derive wantPath from HOME + wantSource
	wantSource   string
	wantErr      bool
}

func TestRuntime_ResolveRepoPath(t *testing.T) {
	cases := []resolveCase{
		{
			name:       "nothing set → embed-fallback",
			wantPath:   "",
			wantSource: "",
		},
		{
			name:       "only flag",
			repoFlag:   "/r",
			wantPath:   "/r",
			wantSource: "flag",
		},
		{
			name:       "only env",
			envValue:   "/e",
			wantPath:   "/e",
			wantSource: "env",
		},
		{
			name: "only config",
			fixture: func(t *testing.T, home string) {
				t.Helper()
				writeConfig(t, home, "repo_path: /c\n")
			},
			wantPath:   "/c",
			wantSource: "config",
		},
		{
			name: "only convention:src",
			fixture: func(t *testing.T, home string) {
				t.Helper()
				mkdirAt(t, filepath.Join(home, "src", "macheim"))
			},
			wantPathFunc: true, // resolved relative to HOME — checked below
			wantSource:   "convention:src",
		},
		{
			name: "only convention:code",
			fixture: func(t *testing.T, home string) {
				t.Helper()
				mkdirAt(t, filepath.Join(home, "code", "macheim"))
			},
			wantPathFunc: true,
			wantSource:   "convention:code",
		},
		{
			name:       "precedence: flag wins over env",
			repoFlag:   "/r",
			envValue:   "/e",
			wantPath:   "/r",
			wantSource: "flag",
		},
		{
			name:     "precedence: env wins over config",
			envValue: "/e",
			fixture: func(t *testing.T, home string) {
				t.Helper()
				writeConfig(t, home, "repo_path: /c\n")
			},
			wantPath:   "/e",
			wantSource: "env",
		},
		{
			name: "precedence: config wins over conventions",
			fixture: func(t *testing.T, home string) {
				t.Helper()
				writeConfig(t, home, "repo_path: /c\n")
				mkdirAt(t, filepath.Join(home, "src", "macheim"))
				mkdirAt(t, filepath.Join(home, "code", "macheim"))
			},
			wantPath:   "/c",
			wantSource: "config",
		},
		{
			name: "precedence: src wins over code",
			fixture: func(t *testing.T, home string) {
				t.Helper()
				mkdirAt(t, filepath.Join(home, "src", "macheim"))
				mkdirAt(t, filepath.Join(home, "code", "macheim"))
			},
			wantPathFunc: true,
			wantSource:   "convention:src",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Tests in this func share env state; do not parallelize.
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("MACHEIM_REPO", tc.envValue)
			if tc.fixture != nil {
				tc.fixture(t, home)
			}
			rt := &Runtime{RepoPath: tc.repoFlag}
			gotPath, gotSource, err := rt.ResolveRepoPath()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want non-nil error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			wantPath := tc.wantPath
			if tc.wantPathFunc {
				// Convention paths are resolved relative to the tempdir HOME.
				switch tc.wantSource {
				case "convention:src":
					wantPath = filepath.Join(home, "src", "macheim")
				case "convention:code":
					wantPath = filepath.Join(home, "code", "macheim")
				default:
					t.Fatalf("wantPathFunc set but wantSource %q is not a convention", tc.wantSource)
				}
			}
			if gotPath != wantPath {
				t.Errorf("path: got %q, want %q", gotPath, wantPath)
			}
			if gotSource != tc.wantSource {
				t.Errorf("source: got %q, want %q", gotSource, tc.wantSource)
			}
		})
	}
}

func TestRuntime_ResolveRepoPath_MalformedConfigErrors(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("MACHEIM_REPO", "")
	writeConfig(t, home, "repo_path: [unterminated\n")

	rt := &Runtime{}
	_, _, err := rt.ResolveRepoPath()
	if err == nil {
		t.Fatalf("want non-nil error for malformed config, got nil")
	}
}

func TestRuntime_IsReadOnly(t *testing.T) {
	t.Run("embed-fallback is read-only", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("MACHEIM_REPO", "")
		rt := &Runtime{}
		if !rt.IsReadOnly() {
			t.Fatalf("want IsReadOnly() == true with no source configured")
		}
	})
	t.Run("flag-configured is not read-only", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("MACHEIM_REPO", "")
		rt := &Runtime{RepoPath: "/r"}
		if rt.IsReadOnly() {
			t.Fatalf("want IsReadOnly() == false when --repo is set")
		}
	})
}

func TestRuntime_VersionString(t *testing.T) {
	t.Parallel()

	rt := &Runtime{
		Version:   "v0.0.1",
		Commit:    "deadbeef",
		BuildDate: "2026-01-01T00:00:00Z",
	}
	got := rt.VersionString()
	want := "v0.0.1 (commit deadbeef, built 2026-01-01T00:00:00Z)"
	if got != want {
		t.Fatalf("VersionString:\n  got:  %q\n  want: %q", got, want)
	}
}

// mkdirAt creates the given directory tree, failing the test on error.
func mkdirAt(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
