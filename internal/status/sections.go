package status

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/gitrepo"
	"github.com/polliard/macheim/internal/output"
)

// brewSeam bundles the OS-touching dependencies of brewRow so tests can
// substitute fakes without hitting the real filesystem or PATH.
type brewSeam struct {
	arch    string
	stat    func(string) (os.FileInfo, error)
	version func(brewPath string) (string, error)
}

func defaultBrewSeam() brewSeam {
	return brewSeam{
		arch: runtime.GOARCH,
		stat: os.Stat,
		version: func(brewPath string) (string, error) {
			out, err := exec.Command(brewPath, "--version").Output() //nolint:gosec // brewPath is one of two canonical paths gated by arch
			if err != nil {
				return "", err
			}
			line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
			parts := strings.Fields(line)
			if len(parts) < 2 {
				return line, nil
			}
			return parts[1], nil
		},
	}
}

// brewSection reports brew presence + path + version.
func brewSection() Section {
	s := defaultBrewSeam()
	return Section{
		Name: "brew",
		Run: func(_ *config.Runtime) []Row {
			return []Row{brewRow(s)}
		},
	}
}

func brewRow(s brewSeam) Row {
	var path string
	switch s.arch {
	case "arm64":
		path = "/opt/homebrew/bin/brew"
	case "amd64":
		path = "/usr/local/bin/brew"
	default:
		return Row{Marker: output.MarkerFail, Name: "brew", Detail: fmt.Sprintf("unsupported arch %q", s.arch)}
	}
	if _, err := s.stat(path); err != nil {
		return Row{Marker: output.MarkerFail, Name: "brew", Detail: "not installed"}
	}
	v, err := s.version(path)
	if err != nil {
		return Row{Marker: output.MarkerOK, Name: "brew", Detail: fmt.Sprintf("%s (version unavailable)", path)}
	}
	return Row{Marker: output.MarkerOK, Name: "brew", Detail: fmt.Sprintf("%s (%s)", path, v)}
}

// repoSeam bundles the dependencies of repoRow so tests can substitute fakes.
type repoSeam struct {
	stat       func(string) (os.FileInfo, error)
	lastCommit func(repoPath string) (sha, subject, isoDate string, err error)
	isClean    func(repoPath string) (bool, error)
}

func defaultRepoSeam() repoSeam {
	return repoSeam{
		stat:       os.Stat,
		lastCommit: gitrepo.LastCommit,
		isClean:    gitrepo.IsClean,
	}
}

// repoSection reports the resolved repo source, path, last commit, and
// clean/dirty state. Embed-fallback mode (no source configured) shows as
// a "?" row, not a failure.
func repoSection() Section {
	s := defaultRepoSeam()
	return Section{
		Name: "repo",
		Run: func(rt *config.Runtime) []Row {
			return []Row{repoRow(rt, s)}
		},
	}
}

func repoRow(rt *config.Runtime, s repoSeam) Row {
	path, source, err := rt.ResolveRepoPath()
	if err != nil {
		return Row{Marker: output.MarkerFail, Name: "repo", Detail: err.Error()}
	}
	if path == "" {
		return Row{
			Marker:  output.MarkerUnknown,
			Name:    "repo",
			Detail:  "not configured (embed-fallback mode)",
			Verbose: "no source matched (--repo flag, MACHEIM_REPO, ~/.config/macheim/config.yaml, ~/src/macheim, ~/code/macheim)",
		}
	}
	if info, err := s.stat(path); err != nil || !info.IsDir() {
		return Row{Marker: output.MarkerFail, Name: "repo", Detail: fmt.Sprintf("[%s] %s (missing)", source, path)}
	}
	sha, subject, isoDate, err := s.lastCommit(path)
	if err != nil {
		return Row{
			Marker:  output.MarkerOK,
			Name:    "repo",
			Detail:  fmt.Sprintf("[%s] %s", source, path),
			Verbose: fmt.Sprintf("git log -1 failed: %v", err),
		}
	}
	clean, _ := s.isClean(path) // IsClean errors fall through as "dirty"
	cleanLabel := "dirty"
	if clean {
		cleanLabel = "clean"
	}
	shortSHA := sha
	if len(sha) >= 7 {
		shortSHA = sha[:7]
	}
	return Row{
		Marker:  output.MarkerOK,
		Name:    "repo",
		Detail:  fmt.Sprintf("[%s] %s @ %s (%s)", source, path, shortSHA, cleanLabel),
		Verbose: fmt.Sprintf("%s — %s", subject, isoDate),
	}
}

// driftSection emits placeholder rows for each module's drift detector,
// pointing at the issues that will land real implementations. Rows are
// Hidden so --quiet collapses them out.
func driftSection() Section {
	return Section{
		Name: "drift",
		Run: func(_ *config.Runtime) []Row {
			return []Row{
				{Marker: output.MarkerUnknown, Name: "drift:brew", Detail: "not implemented (see #14)", Hidden: true},
				{Marker: output.MarkerUnknown, Name: "drift:dotfiles", Detail: "not implemented (see #17 / #18)", Hidden: true},
				{Marker: output.MarkerUnknown, Name: "drift:macos", Detail: "deferred", Hidden: true},
			}
		},
	}
}
