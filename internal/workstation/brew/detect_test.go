//go:build !windows

package brew

import (
	"errors"
	"os"
	"testing"
	"time"
)

// fakeStat is a minimal os.FileInfo stub for table-driven Detect tests.
type fakeStat struct{}

func (fakeStat) Name() string       { return "brew" }
func (fakeStat) Size() int64        { return 0 }
func (fakeStat) Mode() os.FileMode  { return 0 }
func (fakeStat) ModTime() time.Time { return time.Time{} }
func (fakeStat) IsDir() bool        { return false }
func (fakeStat) Sys() any           { return nil }

func TestDetect_ARM64(t *testing.T) {
	t.Parallel()
	s := detectSeam{
		arch: "arm64",
		stat: func(p string) (os.FileInfo, error) {
			if p != "/opt/homebrew/bin/brew" {
				t.Errorf("stat called with %q, want /opt/homebrew/bin/brew", p)
			}
			return fakeStat{}, nil
		},
	}
	path, installed := detect(s)
	if path != "/opt/homebrew/bin/brew" {
		t.Errorf("path = %q, want /opt/homebrew/bin/brew", path)
	}
	if !installed {
		t.Error("installed = false, want true")
	}
}

func TestDetect_AMD64(t *testing.T) {
	t.Parallel()
	s := detectSeam{
		arch: "amd64",
		stat: func(p string) (os.FileInfo, error) {
			if p != "/usr/local/bin/brew" {
				t.Errorf("stat called with %q, want /usr/local/bin/brew", p)
			}
			return fakeStat{}, nil
		},
	}
	path, installed := detect(s)
	if path != "/usr/local/bin/brew" {
		t.Errorf("path = %q, want /usr/local/bin/brew", path)
	}
	if !installed {
		t.Error("installed = false, want true")
	}
}

func TestDetect_NotInstalled(t *testing.T) {
	t.Parallel()
	s := detectSeam{
		arch: "arm64",
		stat: func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
	}
	path, installed := detect(s)
	if path != "/opt/homebrew/bin/brew" {
		t.Errorf("path = %q, want canonical path returned even when missing", path)
	}
	if installed {
		t.Error("installed = true, want false")
	}
}

func TestDetect_Unsupported(t *testing.T) {
	t.Parallel()
	s := detectSeam{
		arch: "riscv64",
		stat: func(string) (os.FileInfo, error) { panic("stat should not be called on unsupported arch") },
	}
	path, installed := detect(s)
	if path != "" {
		t.Errorf("path = %q, want empty string on unsupported arch", path)
	}
	if installed {
		t.Error("installed = true, want false on unsupported arch")
	}
}

func TestVersion_ParsesFirstLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		out  string
		want string
	}{
		{
			name: "multi-line typical output",
			out:  "Homebrew 4.3.5\nHomebrew/homebrew-core (git revision abc; last commit 2026-05-01)\n",
			want: "4.3.5",
		},
		{
			name: "single line no trailing newline",
			out:  "Homebrew 4.0.0",
			want: "4.0.0",
		},
		{
			name: "leading whitespace tolerated",
			out:  "  Homebrew 4.1.2\n",
			want: "4.1.2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := detectSeam{
				run: func(name string, args ...string) (string, error) {
					if name != "/opt/homebrew/bin/brew" {
						t.Errorf("run name = %q, want /opt/homebrew/bin/brew", name)
					}
					if len(args) != 1 || args[0] != "--version" {
						t.Errorf("run args = %v, want [--version]", args)
					}
					return tc.out, nil
				},
			}
			got, err := version(s, "/opt/homebrew/bin/brew")
			if err != nil {
				t.Fatalf("version: %v", err)
			}
			if got != tc.want {
				t.Errorf("version = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestVersion_ExecFails(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("exec: brew: not found")
	s := detectSeam{
		run: func(string, ...string) (string, error) { return "", wantErr },
	}
	got, err := version(s, "/opt/homebrew/bin/brew")
	if err == nil {
		t.Fatal("err = nil, want non-nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want errors.Is wrapping %v", err, wantErr)
	}
	if got != "" {
		t.Errorf("version = %q, want empty string on exec failure", got)
	}
}

func TestVersion_EmptyOutput(t *testing.T) {
	t.Parallel()
	s := detectSeam{
		run: func(string, ...string) (string, error) { return "", nil },
	}
	got, err := version(s, "/opt/homebrew/bin/brew")
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if got != "" {
		t.Errorf("version = %q, want empty string for empty output", got)
	}
}