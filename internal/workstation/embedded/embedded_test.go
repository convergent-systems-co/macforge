//go:build !windows

package embedded

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveFile_RepoPreferred asserts that when repoPath is set and
// the named file exists under it, ResolveFile returns the repo path
// with source = "repo" — the on-disk repo always wins over the embed.
func TestResolveFile_RepoPreferred(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	repoBrewfile := filepath.Join(dir, "Brewfile")
	if err := os.WriteFile(repoBrewfile, []byte("brew \"git\"\n"), 0o644); err != nil {
		t.Fatalf("seed repo Brewfile: %v", err)
	}

	path, source, err := ResolveFile(dir, "Brewfile")
	if err != nil {
		t.Fatalf("ResolveFile returned error: %v", err)
	}
	if source != SourceRepo {
		t.Errorf("source = %q, want %q", source, SourceRepo)
	}
	if path != repoBrewfile {
		t.Errorf("path = %q, want %q", path, repoBrewfile)
	}
}

// TestResolveFile_EmbedFallback_Missing asserts that when neither the
// repo nor the embed contains the requested name, ResolveFile returns
// a descriptive error rather than a silent empty result.
func TestResolveFile_EmbedFallback_Missing(t *testing.T) {
	t.Parallel()

	path, source, err := ResolveFile("", "definitely-not-a-real-asset.txt")
	if err == nil {
		t.Fatalf("expected error, got path=%q source=%q", path, source)
	}
	if !strings.Contains(err.Error(), "definitely-not-a-real-asset.txt") {
		t.Errorf("error %q does not mention the missing name", err.Error())
	}
	if path != "" || source != "" {
		t.Errorf("on error want empty path/source, got path=%q source=%q", path, source)
	}
}

// TestResolveFile_EmbedFallback_Present asserts that when repoPath is
// empty (no clone yet) but the embed contains the requested name,
// ResolveFile returns the embed FS path with source = "embed".
func TestResolveFile_EmbedFallback_Present(t *testing.T) {
	t.Parallel()

	path, source, err := ResolveFile("", "Brewfile")
	if err != nil {
		t.Fatalf("ResolveFile returned error: %v", err)
	}
	if source != SourceEmbed {
		t.Errorf("source = %q, want %q", source, SourceEmbed)
	}
	if path != "configs/Brewfile" {
		t.Errorf("path = %q, want %q", path, "configs/Brewfile")
	}
	// Sanity: the returned embed path is actually readable via Configs.
	if _, err := Configs.ReadFile(path); err != nil {
		t.Errorf("Configs.ReadFile(%q): %v", path, err)
	}
}