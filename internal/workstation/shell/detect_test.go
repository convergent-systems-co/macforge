//go:build !windows

package shell

import (
	"path/filepath"
	"testing"
)

func TestDetect(t *testing.T) {
	// Subtests are NOT run in parallel: each mutates SHELL and HOME via
	// t.Setenv, which the testing package serializes per-test but does
	// not isolate across parallel siblings.
	home := "/tmp/macheim-detect-home"

	tests := []struct {
		name      string
		shell     string
		wantShell string
		wantRC    string
		wantErr   bool
	}{
		{name: "zsh from /bin/zsh", shell: "/bin/zsh", wantShell: "zsh", wantRC: filepath.Join(home, ".zshrc")},
		{name: "zsh from homebrew path", shell: "/opt/homebrew/bin/zsh", wantShell: "zsh", wantRC: filepath.Join(home, ".zshrc")},
		{name: "bash from /bin/bash", shell: "/bin/bash", wantShell: "bash", wantRC: filepath.Join(home, ".bash_profile")},
		{name: "fish is unknown", shell: "/usr/bin/fish", wantErr: true},
		{name: "empty SHELL is unknown", shell: "", wantErr: true},
		{name: "exotic shell is unknown", shell: "/opt/strange/nushell", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SHELL", tt.shell)
			t.Setenv("HOME", home)

			sh, rc, err := Detect()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for SHELL=%q, got (%q, %q)", tt.shell, sh, rc)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sh != tt.wantShell {
				t.Errorf("shell: got %q want %q", sh, tt.wantShell)
			}
			if rc != tt.wantRC {
				t.Errorf("rcPath: got %q want %q", rc, tt.wantRC)
			}
		})
	}
}