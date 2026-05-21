package status

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/output"
)

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

func TestBrewRow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		arch       string
		stat       func(string) (os.FileInfo, error)
		version    func(string) (string, error)
		wantMarker output.Marker
		wantDetail string
	}{
		{
			name:       "arm64 happy",
			arch:       "arm64",
			stat:       func(string) (os.FileInfo, error) { return fakeStat{}, nil },
			version:    func(string) (string, error) { return "4.3.5", nil },
			wantMarker: output.MarkerOK,
			wantDetail: "/opt/homebrew/bin/brew (4.3.5)",
		},
		{
			name:       "amd64 happy",
			arch:       "amd64",
			stat:       func(string) (os.FileInfo, error) { return fakeStat{}, nil },
			version:    func(string) (string, error) { return "4.3.5", nil },
			wantMarker: output.MarkerOK,
			wantDetail: "/usr/local/bin/brew (4.3.5)",
		},
		{
			name:       "missing arm64",
			arch:       "arm64",
			stat:       func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			version:    func(string) (string, error) { return "", nil },
			wantMarker: output.MarkerFail,
			wantDetail: "not installed",
		},
		{
			name:       "version exec fails",
			arch:       "arm64",
			stat:       func(string) (os.FileInfo, error) { return fakeStat{}, nil },
			version:    func(string) (string, error) { return "", errors.New("oops") },
			wantMarker: output.MarkerOK,
			wantDetail: "/opt/homebrew/bin/brew (version unavailable)",
		},
		{
			name:       "unsupported arch",
			arch:       "riscv64",
			wantMarker: output.MarkerFail,
			wantDetail: `unsupported arch "riscv64"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s := brewSeam{arch: tc.arch, stat: tc.stat, version: tc.version}
			got := brewRow(s)
			if got.Marker != tc.wantMarker {
				t.Errorf("marker: got %v, want %v (detail=%q)", got.Marker, tc.wantMarker, got.Detail)
			}
			if got.Detail != tc.wantDetail {
				t.Errorf("detail: got %q, want %q", got.Detail, tc.wantDetail)
			}
		})
	}
}

func TestRepoRow(t *testing.T) {
	// Cannot t.Parallel — subtests call t.Setenv to clear MACHEIM_REPO and HOME.
	tests := []struct {
		name        string
		repoFlag    string
		stat        func(string) (os.FileInfo, error)
		lastCommit  func(string) (string, string, string, error)
		isClean     func(string) (bool, error)
		wantMarker  output.Marker
		wantContain []string
	}{
		{
			name:        "embed fallback (nothing configured)",
			repoFlag:    "",
			wantMarker:  output.MarkerUnknown,
			wantContain: []string{"not configured", "embed-fallback"},
		},
		{
			name:     "configured + clean + has commit",
			repoFlag: "/r",
			stat:     func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			lastCommit: func(string) (string, string, string, error) {
				return "abc1234deadbeef", "subj", "2026-05-20T00:00:00Z", nil
			},
			isClean:     func(string) (bool, error) { return true, nil },
			wantMarker:  output.MarkerOK,
			wantContain: []string{"[flag]", "/r", "abc1234", "clean"},
		},
		{
			name:     "configured + dirty",
			repoFlag: "/r",
			stat:     func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			lastCommit: func(string) (string, string, string, error) {
				return "abc1234deadbeef", "subj", "2026-05-20T00:00:00Z", nil
			},
			isClean:     func(string) (bool, error) { return false, nil },
			wantMarker:  output.MarkerOK,
			wantContain: []string{"dirty"},
		},
		{
			name:        "configured but missing on disk",
			repoFlag:    "/missing",
			stat:        func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			wantMarker:  output.MarkerFail,
			wantContain: []string{"[flag]", "/missing", "missing"},
		},
		{
			name:        "lastCommit fails — degraded but ok",
			repoFlag:    "/r",
			stat:        func(string) (os.FileInfo, error) { return fakeStat{isDir: true}, nil },
			lastCommit:  func(string) (string, string, string, error) { return "", "", "", errors.New("not a git repo") },
			wantMarker:  output.MarkerOK,
			wantContain: []string{"[flag]", "/r"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MACHEIM_REPO", "")
			t.Setenv("HOME", t.TempDir()) // isolate config.Load() + conventions
			rt := &config.Runtime{RepoPath: tc.repoFlag}
			s := repoSeam{stat: tc.stat, lastCommit: tc.lastCommit, isClean: tc.isClean}
			got := repoRow(rt, s)
			if got.Marker != tc.wantMarker {
				t.Errorf("marker: got %v, want %v (detail=%q)", got.Marker, tc.wantMarker, got.Detail)
			}
			for _, want := range tc.wantContain {
				if !strings.Contains(got.Detail, want) {
					t.Errorf("detail %q does not contain %q", got.Detail, want)
				}
			}
		})
	}
}
