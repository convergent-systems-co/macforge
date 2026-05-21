package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// xcodeCheck verifies `xcode-select -p` succeeds and names a directory.
func xcodeCheck(_ *config.Runtime, s seam) Result {
	out, err := s.run("xcode-select", "-p")
	if err != nil {
		return Result{
			Probe:       "ran `xcode-select -p`: command failed",
			Remediation: "Run: xcode-select --install",
		}
	}
	path := strings.TrimSpace(out)
	if info, err := s.stat(path); err != nil || !info.IsDir() {
		return Result{
			Probe:       fmt.Sprintf("xcode-select -p → %s (path missing)", path),
			Remediation: "Run: xcode-select --install",
		}
	}
	return Result{OK: true, Probe: fmt.Sprintf("xcode-select -p → %s", path)}
}

// brewCheck verifies the arch-appropriate brew binary exists.
func brewCheck(_ *config.Runtime, s seam) Result {
	var path string
	switch s.arch {
	case "arm64":
		path = "/opt/homebrew/bin/brew"
	case "amd64":
		path = "/usr/local/bin/brew"
	default:
		return Result{
			Probe:       fmt.Sprintf("unsupported architecture %q", s.arch),
			Remediation: "macheim runs on arm64 (Apple Silicon) and amd64 (Intel) only",
		}
	}
	if _, err := s.stat(path); err != nil {
		return Result{
			Probe:       fmt.Sprintf("%s (not found)", path),
			Remediation: "Run: macheim brew install",
		}
	}
	return Result{OK: true, Probe: path}
}

// repoCheck verifies the resolved repo path exists and is writable, OR
// reports the embed-fallback case (no source configured) as a pass.
func repoCheck(rt *config.Runtime, s seam) Result {
	path, source, err := rt.ResolveRepoPath()
	if err != nil {
		return Result{Probe: err.Error(), Remediation: "Set --repo or MACHEIM_REPO"}
	}
	if path == "" {
		return Result{OK: true, Probe: "no repo configured; running in embed-fallback mode"}
	}
	info, err := s.stat(path)
	if err != nil || !info.IsDir() {
		return Result{
			Probe:       fmt.Sprintf("[%s] %s (missing)", source, path),
			Remediation: fmt.Sprintf("Clone the macheim repo to %s, or update --repo / MACHEIM_REPO", path),
		}
	}
	if !s.canWriteDir(path) {
		return Result{
			Probe:       fmt.Sprintf("[%s] %s (not writable)", source, path),
			Remediation: fmt.Sprintf("Check permissions on %s", path),
		}
	}
	return Result{OK: true, Probe: fmt.Sprintf("[%s] %s", source, path)}
}

// configDirCheck verifies ~/.config/macheim/ exists and is writable, OR
// is missing but the parent ~/.config/ is writable (so we can create it).
func configDirCheck(_ *config.Runtime, s seam) Result {
	cfgDir := filepath.Join(s.homeDir, ".config", "macheim")
	if _, err := s.stat(cfgDir); err == nil {
		if !s.canWriteDir(cfgDir) {
			return Result{
				Probe:       fmt.Sprintf("%s (exists, not writable)", cfgDir),
				Remediation: fmt.Sprintf("chmod u+w %s", cfgDir),
			}
		}
		return Result{OK: true, Probe: fmt.Sprintf("%s (writable)", cfgDir)}
	}
	parent := filepath.Dir(cfgDir)
	if s.canWriteDir(parent) {
		return Result{OK: true, Probe: fmt.Sprintf("%s (will create; parent %s writable)", cfgDir, parent)}
	}
	return Result{
		Probe:       fmt.Sprintf("%s (missing; parent %s not writable)", cfgDir, parent),
		Remediation: fmt.Sprintf("mkdir -p %s && chmod u+w %s", cfgDir, cfgDir),
	}
}

// shellRCCheck verifies $SHELL names zsh or bash AND the appropriate rc
// file is writable (or creatable when missing).
func shellRCCheck(_ *config.Runtime, s seam) Result {
	sh := s.lookEnv("SHELL")
	var rc string
	switch {
	case strings.HasSuffix(sh, "zsh"):
		rc = filepath.Join(s.homeDir, ".zshrc")
	case strings.HasSuffix(sh, "bash"):
		rc = filepath.Join(s.homeDir, ".bash_profile")
	default:
		return Result{
			Probe:       fmt.Sprintf("$SHELL=%q (unknown)", sh),
			Remediation: "Set SHELL to /bin/zsh or /bin/bash",
		}
	}
	if s.canWriteFile(rc) {
		return Result{OK: true, Probe: fmt.Sprintf("$SHELL=%s → %s (writable)", sh, rc)}
	}
	return Result{
		Probe:       fmt.Sprintf("$SHELL=%s → %s (not writable)", sh, rc),
		Remediation: fmt.Sprintf("touch %s && chmod u+w %s", rc, rc),
	}
}

// dirWritable returns true iff dir exists and the caller can create files
// in it. Probes by creating then removing a temp directory.
func dirWritable(dir string) bool {
	probe, err := os.MkdirTemp(dir, "macheim-doctor-")
	if err != nil {
		return false
	}
	_ = os.Remove(probe)
	return true
}

// fileOrParentWritable returns true when the path is writable, or when
// the path doesn't exist but its parent directory is writable (so we
// could create it). Existing files are probed via O_WRONLY|O_APPEND so
// no bytes are written.
func fileOrParentWritable(path string) bool {
	info, err := os.Stat(path)
	if err == nil && !info.IsDir() {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			return false
		}
		_ = f.Close()
		return true
	}
	if os.IsNotExist(err) {
		return dirWritable(filepath.Dir(path))
	}
	return false
}
