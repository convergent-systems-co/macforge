package doctor

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// fakeStat returns a stub os.FileInfo for table tests.
type fakeStat struct {
	name  string
	isDir bool
}

func (f fakeStat) Name() string       { return f.name }
func (f fakeStat) Size() int64        { return 0 }
func (f fakeStat) Mode() os.FileMode  { return 0 }
func (f fakeStat) ModTime() time.Time { return time.Time{} }
func (f fakeStat) IsDir() bool        { return f.isDir }
func (f fakeStat) Sys() any           { return nil }

// makeSeam builds a seam from a set of overrides. Anything not set
// falls back to a panicking default so accidental real-OS calls in
// tests are loud.
func makeSeam(overrides seam) seam {
	if overrides.run == nil {
		overrides.run = func(name string, args ...string) (string, error) {
			panic("unstubbed seam.run call: " + name)
		}
	}
	if overrides.stat == nil {
		overrides.stat = func(p string) (os.FileInfo, error) { panic("unstubbed seam.stat call: " + p) }
	}
	if overrides.lookEnv == nil {
		overrides.lookEnv = func(string) string { return "" }
	}
	if overrides.canWriteDir == nil {
		overrides.canWriteDir = func(string) bool { return true }
	}
	if overrides.canWriteFile == nil {
		overrides.canWriteFile = func(string) bool { return true }
	}
	if overrides.arch == "" {
		overrides.arch = "arm64"
	}
	if overrides.homeDir == "" {
		overrides.homeDir = "/home/test"
	}
	return overrides
}

func TestXcodeCheck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		run     func(string, ...string) (string, error)
		stat    func(string) (os.FileInfo, error)
		wantOK  bool
		wantRem string
	}{
		{
			name:   "happy",
			run:    func(string, ...string) (string, error) { return "/path\n", nil },
			stat:   func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			wantOK: true,
		},
		{
			name:    "command fails",
			run:     func(string, ...string) (string, error) { return "", errors.New("not found") },
			stat:    func(string) (os.FileInfo, error) { panic("stat should not be called") },
			wantOK:  false,
			wantRem: "xcode-select --install",
		},
		{
			name:    "path missing",
			run:     func(string, ...string) (string, error) { return "/p\n", nil },
			stat:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			wantOK:  false,
			wantRem: "xcode-select --install",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := xcodeCheck(&config.Runtime{}, makeSeam(seam{run: tc.run, stat: tc.stat}))
			if r.OK != tc.wantOK {
				t.Errorf("OK: got %v, want %v (probe=%q)", r.OK, tc.wantOK, r.Probe)
			}
			if tc.wantRem != "" && !strings.Contains(r.Remediation, tc.wantRem) {
				t.Errorf("Remediation %q does not contain %q", r.Remediation, tc.wantRem)
			}
		})
	}
}

func TestBrewCheck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		arch     string
		stat     func(string) (os.FileInfo, error)
		wantOK   bool
		wantPath string
		wantRem  string
	}{
		{
			name:     "arm64 happy",
			arch:     "arm64",
			stat:     func(p string) (os.FileInfo, error) { return fakeStat{}, nil },
			wantOK:   true,
			wantPath: "/opt/homebrew/bin/brew",
		},
		{
			name:     "amd64 happy",
			arch:     "amd64",
			stat:     func(p string) (os.FileInfo, error) { return fakeStat{}, nil },
			wantOK:   true,
			wantPath: "/usr/local/bin/brew",
		},
		{
			name:    "missing on arm64",
			arch:    "arm64",
			stat:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			wantOK:  false,
			wantRem: "macheim brew install",
		},
		{
			name:    "unsupported arch",
			arch:    "riscv64",
			stat:    func(string) (os.FileInfo, error) { panic("stat should not be called") },
			wantOK:  false,
			wantRem: "arm64",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := brewCheck(&config.Runtime{}, makeSeam(seam{arch: tc.arch, stat: tc.stat}))
			if r.OK != tc.wantOK {
				t.Errorf("OK: got %v, want %v (probe=%q)", r.OK, tc.wantOK, r.Probe)
			}
			if tc.wantPath != "" && !strings.Contains(r.Probe, tc.wantPath) {
				t.Errorf("Probe %q does not contain %q", r.Probe, tc.wantPath)
			}
			if tc.wantRem != "" && !strings.Contains(r.Remediation, tc.wantRem) {
				t.Errorf("Remediation %q does not contain %q", r.Remediation, tc.wantRem)
			}
		})
	}
}

func TestRepoCheck(t *testing.T) {
	// Cannot t.Parallel — subtests call t.Setenv to clear MACHEIM_REPO,
	// which panics when any ancestor test is in parallel mode.
	tests := []struct {
		name        string
		repoPath    string
		stat        func(string) (os.FileInfo, error)
		canWriteDir func(string) bool
		wantOK      bool
		wantProbe   string
		wantRem     string
	}{
		{
			name:      "no source — embed fallback",
			repoPath:  "",
			wantOK:    true,
			wantProbe: "embed-fallback",
		},
		{
			name:        "configured + writable",
			repoPath:    "/repo",
			stat:        func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			canWriteDir: func(string) bool { return true },
			wantOK:      true,
			wantProbe:   "/repo",
		},
		{
			name:      "configured + missing",
			repoPath:  "/missing",
			stat:      func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			wantOK:    false,
			wantProbe: "missing",
			wantRem:   "Clone the macheim repo",
		},
		{
			name:        "configured + not writable",
			repoPath:    "/locked",
			stat:        func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			canWriteDir: func(string) bool { return false },
			wantOK:      false,
			wantProbe:   "not writable",
			wantRem:     "permissions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// MACHEIM_REPO is set via t.Setenv; can't run repo subtests in parallel
			// because they all share env state — table runs sequentially.
			t.Setenv("MACHEIM_REPO", "")
			rt := &config.Runtime{RepoPath: tc.repoPath}
			r := repoCheck(rt, makeSeam(seam{stat: tc.stat, canWriteDir: tc.canWriteDir}))
			if r.OK != tc.wantOK {
				t.Errorf("OK: got %v, want %v (probe=%q)", r.OK, tc.wantOK, r.Probe)
			}
			if tc.wantProbe != "" && !strings.Contains(r.Probe, tc.wantProbe) {
				t.Errorf("Probe %q does not contain %q", r.Probe, tc.wantProbe)
			}
			if tc.wantRem != "" && !strings.Contains(r.Remediation, tc.wantRem) {
				t.Errorf("Remediation %q does not contain %q", r.Remediation, tc.wantRem)
			}
		})
	}
}

func TestConfigDirCheck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		stat        func(string) (os.FileInfo, error)
		canWriteDir func(string) bool
		wantOK      bool
	}{
		{
			name:        "exists writable",
			stat:        func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			canWriteDir: func(string) bool { return true },
			wantOK:      true,
		},
		{
			name:        "exists not writable",
			stat:        func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			canWriteDir: func(string) bool { return false },
			wantOK:      false,
		},
		{
			name:        "missing, parent writable",
			stat:        func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			canWriteDir: func(p string) bool { return strings.HasSuffix(p, "/.config") },
			wantOK:      true,
		},
		{
			name:        "missing, parent not writable",
			stat:        func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			canWriteDir: func(string) bool { return false },
			wantOK:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := configDirCheck(&config.Runtime{}, makeSeam(seam{stat: tc.stat, canWriteDir: tc.canWriteDir}))
			if r.OK != tc.wantOK {
				t.Errorf("OK: got %v, want %v (probe=%q)", r.OK, tc.wantOK, r.Probe)
			}
		})
	}
}

func TestShellRCCheck(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		shell        string
		canWriteFile func(string) bool
		wantOK       bool
		wantProbe    string
		wantRem      string
	}{
		{name: "zsh writable", shell: "/bin/zsh", canWriteFile: func(string) bool { return true }, wantOK: true, wantProbe: ".zshrc"},
		{name: "bash writable", shell: "/bin/bash", canWriteFile: func(string) bool { return true }, wantOK: true, wantProbe: ".bash_profile"},
		{name: "zsh not writable", shell: "/bin/zsh", canWriteFile: func(string) bool { return false }, wantOK: false, wantRem: "touch"},
		{name: "unknown shell", shell: "/bin/fish", wantOK: false, wantRem: "Set SHELL"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := makeSeam(seam{
				lookEnv:      func(string) string { return tc.shell },
				canWriteFile: tc.canWriteFile,
			})
			r := shellRCCheck(&config.Runtime{}, s)
			if r.OK != tc.wantOK {
				t.Errorf("OK: got %v, want %v (probe=%q)", r.OK, tc.wantOK, r.Probe)
			}
			if tc.wantProbe != "" && !strings.Contains(r.Probe, tc.wantProbe) {
				t.Errorf("Probe %q does not contain %q", r.Probe, tc.wantProbe)
			}
			if tc.wantRem != "" && !strings.Contains(r.Remediation, tc.wantRem) {
				t.Errorf("Remediation %q does not contain %q", r.Remediation, tc.wantRem)
			}
		})
	}
}

// Smoke test for the writability helpers — exercises real filesystem.
func TestWritabilityHelpers_RealFS(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if !dirWritable(dir) {
		t.Error("freshly-created TempDir should be writable")
	}
	if dirWritable("/this/path/does/not/exist/anywhere") {
		t.Error("nonexistent dir reported writable")
	}

	// File that doesn't exist yet — parent (TempDir) is writable.
	missing := dir + "/not-yet.txt"
	if !fileOrParentWritable(missing) {
		t.Error("nonexistent file in writable dir should be reported writable")
	}

	// Existing file with write permission.
	existing := dir + "/exists.txt"
	if err := os.WriteFile(existing, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !fileOrParentWritable(existing) {
		t.Error("existing 0600 file should be writable by owner")
	}

	// Existing file with no write permission.
	readOnly := dir + "/ro.txt"
	if err := os.WriteFile(readOnly, []byte("x"), 0o400); err != nil {
		t.Fatal(err)
	}
	if fileOrParentWritable(readOnly) {
		t.Error("0400 file should not be reported writable")
	}
}
