package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Detect inspects $SHELL and returns the shell's short name and the
// canonical rc-file path under $HOME:
//
//   - $SHELL ends in "zsh"  → ("zsh",  $HOME/.zshrc, nil)
//   - $SHELL ends in "bash" → ("bash", $HOME/.bash_profile, nil)
//   - anything else         → ("", "", error)
//
// The bash convention is intentionally .bash_profile (not .bashrc):
// macOS Terminal.app and iTerm2 both run login shells by default, so
// .bash_profile is what gets sourced. Linux desktop terminals tend to
// run non-login shells and prefer .bashrc — macheim is a macOS tool, so
// .bash_profile is correct here.
//
// Suffix matching tolerates any prefix (/bin/zsh, /opt/homebrew/bin/zsh,
// /usr/local/Cellar/bash/5.2/bin/bash, etc.) without committing to a
// specific install path.
func Detect() (shell, rcPath string, err error) {
	sh := os.Getenv("SHELL")
	home := os.Getenv("HOME")

	switch {
	case strings.HasSuffix(sh, "zsh"):
		return "zsh", filepath.Join(home, ".zshrc"), nil
	case strings.HasSuffix(sh, "bash"):
		return "bash", filepath.Join(home, ".bash_profile"), nil
	default:
		return "", "", fmt.Errorf("unknown shell %q", sh)
	}
}
