// Package embedded exposes static files baked into the macheim binary at
// build time. The companion `make embed-sync` target copies the source
// files from the repo into this package before `go build`, so a fresh
// Mac with no repo cloned can still bootstrap from just the binary.
//
// At runtime, ResolveFile implements the repo-vs-embedded precedence
// described in GOALS.md: when a repo path is known and contains the
// requested file, that on-disk copy wins; otherwise we fall back to the
// embedded snapshot. Embedded files are read-only — write paths must
// require a repo.
package embedded

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Scripts holds shell scripts shipped with the binary.
//
// install-brew.sh is a pinned snapshot of Homebrew's official installer;
// see the script header for the refresh procedure.
//
//go:embed scripts/install-brew.sh
var Scripts embed.FS

// Configs holds config-file fallbacks (Brewfile, dotfiles tree, etc.)
// populated by `make embed-sync` from the repo's own working copy.
//
//go:embed configs
var Configs embed.FS

// Source identifies where ResolveFile found a file.
const (
	SourceRepo  = "repo"
	SourceEmbed = "embed"
)

// ResolveFile returns the location to read for the named asset
// (for example "Brewfile" or "dotfiles").
//
// Precedence:
//  1. If repoPath is non-empty and "<repoPath>/<name>" exists on disk,
//     return that absolute path with source = "repo". Callers read it
//     with os.ReadFile / os.Open.
//  2. Otherwise, if "configs/<name>" exists inside the embedded Configs
//     FS, return that FS-relative path with source = "embed". Callers
//     read it via Configs.ReadFile / fs.Sub.
//  3. Otherwise, return a descriptive error.
//
// ResolveFile does not read file contents — it only decides where to
// read from. This keeps it cheap and lets callers stream large files.
func ResolveFile(repoPath, name string) (path, source string, err error) {
	if repoPath != "" {
		candidate := filepath.Join(repoPath, name)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate, SourceRepo, nil
		}
	}

	embedPath := "configs/" + name
	if _, statErr := fs.Stat(Configs, embedPath); statErr == nil {
		return embedPath, SourceEmbed, nil
	}

	return "", "", fmt.Errorf("file %q not found in repo or embedded", name)
}
