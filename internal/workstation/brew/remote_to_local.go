package brew

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
	"github.com/convergent-systems-co/macforge/internal/workstation/embedded"
	"github.com/convergent-systems-co/macforge/internal/workstation/gitrepo"
	"github.com/convergent-systems-co/macforge/internal/workstation/shell"
)

// remoteToLocalSeam bundles the host-touching dependencies that
// UpdateRemoteToLocal needs. Tests substitute fakes for detect,
// run, gitrepo.IsClean, and gitrepo.Pull. The zero value is unusable;
// production code uses defaultRemoteToLocalSeam.
type remoteToLocalSeam struct {
	detect    func() (path string, installed bool)
	run       func(rt *config.Runtime, name string, args ...string) (string, error)
	isClean   func(repoPath string) (bool, error)
	pull      func(rt *config.Runtime, repoPath string) error
	resolveFn func(repoPath, name string) (path, source string, err error)
}

func defaultRemoteToLocalSeam() remoteToLocalSeam {
	return remoteToLocalSeam{
		detect:    Detect,
		run:       shell.Run,
		isClean:   gitrepo.IsClean,
		pull:      gitrepo.Pull,
		resolveFn: embedded.ResolveFile,
	}
}

// UpdateRemoteToLocal applies the repo Brewfile to this Mac. Steps:
//
//  1. Resolve the repo path. In embed-fallback mode (no repo) we skip
//     the pull and dirty-tree checks and run `brew bundle` against the
//     embedded Brewfile fallback — the closest thing to "apply the
//     known-good state" when no repo has been cloned.
//  2. If a repo is configured and !noPull, refuse when the working
//     tree has uncommitted changes (the pull would clobber them);
//     otherwise `git -C <repo> pull --ff-only`.
//  3. Detect brew; refuse if not installed.
//  4. Run `<brew> bundle --file=<brewfilePath>`; with prune, append
//     --cleanup so brew bundle removes formulae no longer listed.
//
// This implementation does not import internal/brew.Apply (the TL that
// owns brew bundle wraps it separately) — the brew-bundle invocation
// is inlined here so the file is independent of that parallel work.
func UpdateRemoteToLocal(rt *config.Runtime, prune, noPull bool) error {
	return updateRemoteToLocal(defaultRemoteToLocalSeam(), rt, prune, noPull)
}

func updateRemoteToLocal(s remoteToLocalSeam, rt *config.Runtime, prune, noPull bool) error {
	repoPath, _, err := rt.ResolveRepoPath()
	if err != nil {
		return err
	}

	brewfilePath, err := resolveBrewfilePath(s, repoPath)
	if err != nil {
		return err
	}

	if repoPath != "" && !noPull {
		if err := pullIfClean(s, rt, repoPath); err != nil {
			return err
		}
	}

	brewPath, installed := s.detect()
	if !installed {
		return errors.New("brew not installed; run `macheim brew install` first")
	}

	args := []string{"bundle", "--file=" + brewfilePath}
	if prune {
		args = append(args, "--cleanup")
	}
	if _, err := s.run(rt, brewPath, args...); err != nil {
		return fmt.Errorf("brew bundle: %w", err)
	}
	return nil
}

// resolveBrewfilePath returns the on-disk path to the Brewfile we will
// hand to `brew bundle`. When a repo is configured we always use
// <repo>/Brewfile (even if it doesn't exist yet — brew bundle will
// produce its own clean error). When no repo is configured we fall
// back to the embedded snapshot, which embedded.ResolveFile materializes
// as an embed-FS-relative path. Since brew bundle needs a real file, we
// surface a clean error in that case directing the user to clone the
// repo first.
//
// We don't extract the embedded Brewfile to a tempfile here because the
// embed-fallback path for brew bundle isn't yet a supported workflow —
// users in embed-fallback mode should run `macheim bootstrap` to clone
// the repo first. Mentioned in GOALS.md as the standard onboarding flow.
func resolveBrewfilePath(s remoteToLocalSeam, repoPath string) (string, error) {
	if repoPath != "" {
		return filepath.Join(repoPath, "Brewfile"), nil
	}
	path, source, err := s.resolveFn(repoPath, "Brewfile")
	if err != nil {
		return "", fmt.Errorf("locate Brewfile: %w", err)
	}
	if source == embedded.SourceEmbed {
		return "", errors.New("update remote-to-local: no repo configured; clone the macheim repo and set --repo or MACHEIM_REPO first (the embedded Brewfile is read-only)")
	}
	return path, nil
}

// pullIfClean is the dirty-tree guard + git pull bundle. The guard
// runs *before* the pull so a user with uncommitted changes is told
// what to do instead of having the merge fail mid-way.
func pullIfClean(s remoteToLocalSeam, rt *config.Runtime, repoPath string) error {
	clean, err := s.isClean(repoPath)
	if err != nil {
		return fmt.Errorf("check repo cleanliness: %w", err)
	}
	if !clean {
		return fmt.Errorf("update remote-to-local: repo at %s has uncommitted changes; commit or stash before pulling", repoPath)
	}
	if err := s.pull(rt, repoPath); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}
