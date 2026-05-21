package brew

import (
	"errors"
	"strings"
	"testing"

	"github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// stubInstallSeam builds an installSeam from overrides. Any field left
// nil falls back to a panicking default so accidental real-host calls
// in tests are loud rather than silent.
func stubInstallSeam(overrides installSeam) installSeam {
	if overrides.detect == nil {
		overrides.detect = func() (string, bool) { panic("unstubbed detect call") }
	}
	if overrides.scripts == nil {
		overrides.scripts = func() ([]byte, error) { panic("unstubbed scripts call") }
	}
	if overrides.tempFile == nil {
		overrides.tempFile = func([]byte) (string, func(), error) { panic("unstubbed tempFile call") }
	}
	if overrides.run == nil {
		overrides.run = func(*config.Runtime, string, ...string) (string, error) { panic("unstubbed run call") }
	}
	if overrides.shellDetect == nil {
		overrides.shellDetect = func() (string, string, error) { return "zsh", "/home/test/.zshrc", nil }
	}
	if overrides.appendIfMissing == nil {
		overrides.appendIfMissing = func(*config.Runtime, string, string) error { return nil }
	}
	if overrides.version == nil {
		overrides.version = func(string) (string, error) { return "4.3.5", nil }
	}
	if overrides.stderr == nil {
		overrides.stderr = func(string, ...any) {}
	}
	return overrides
}

func TestInstall_AlreadyInstalled(t *testing.T) {
	t.Parallel()
	var (
		runCalled    bool
		appendCalled bool
		appendedLine string
	)
	s := stubInstallSeam(installSeam{
		detect: func() (string, bool) { return "/opt/homebrew/bin/brew", true },
		run: func(*config.Runtime, string, ...string) (string, error) {
			runCalled = true
			return "", nil
		},
		appendIfMissing: func(_ *config.Runtime, rcPath, line string) error {
			appendCalled = true
			appendedLine = line
			if rcPath != "/home/test/.zshrc" {
				t.Errorf("appendIfMissing rcPath = %q, want /home/test/.zshrc", rcPath)
			}
			return nil
		},
	})

	if err := install(s, &config.Runtime{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	if runCalled {
		t.Error("shell.Run called when brew already installed; expected skip")
	}
	if !appendCalled {
		t.Error("appendIfMissing not called; expected idempotent shellenv append even when brew was preinstalled")
	}
	if appendedLine != shellenvLine {
		t.Errorf("appended line = %q, want %q", appendedLine, shellenvLine)
	}
}

// freshInstallSpy captures the seam callbacks for TestInstall_FreshInstall.
type freshInstallSpy struct {
	detectCalls    int
	scriptsCalled  bool
	tempFileCalled bool
	cleanupCalled  bool
	appendCalled   bool
	versionCalled  bool
	runName        string
	runArgs        []string
}

func newFreshInstallSeam(t *testing.T, spy *freshInstallSpy) installSeam {
	t.Helper()
	return stubInstallSeam(installSeam{
		detect: func() (string, bool) {
			spy.detectCalls++
			// First call: not installed. Second call (post-install
			// re-detect): installed.
			return "/opt/homebrew/bin/brew", spy.detectCalls > 1
		},
		scripts: func() ([]byte, error) {
			spy.scriptsCalled = true
			return []byte("#!/bin/bash\necho fake installer\n"), nil
		},
		tempFile: func(content []byte) (string, func(), error) {
			spy.tempFileCalled = true
			if !strings.Contains(string(content), "fake installer") {
				t.Errorf("tempFile content does not contain script: %q", string(content))
			}
			return "/tmp/macheim-fake.sh", func() { spy.cleanupCalled = true }, nil
		},
		run: func(_ *config.Runtime, name string, args ...string) (string, error) {
			spy.runName = name
			spy.runArgs = args
			return "", nil
		},
		appendIfMissing: func(*config.Runtime, string, string) error {
			spy.appendCalled = true
			return nil
		},
		version: func(string) (string, error) {
			spy.versionCalled = true
			return "4.3.5", nil
		},
	})
}

func assertFreshInstallSpy(t *testing.T, spy *freshInstallSpy) {
	t.Helper()
	assertFreshInstallCalls(t, spy)
	assertFreshInstallRunArgs(t, spy.runName, spy.runArgs)
	if spy.detectCalls != 2 {
		t.Errorf("detect called %d times, want 2 (initial + post-install)", spy.detectCalls)
	}
}

func assertFreshInstallCalls(t *testing.T, spy *freshInstallSpy) {
	t.Helper()
	if !spy.scriptsCalled {
		t.Error("scripts not loaded")
	}
	if !spy.tempFileCalled {
		t.Error("tempFile not created")
	}
	if !spy.cleanupCalled {
		t.Error("temp file cleanup not invoked")
	}
	if !spy.appendCalled {
		t.Error("appendIfMissing not called after install")
	}
	if !spy.versionCalled {
		t.Error("version verification not called after install")
	}
}

func assertFreshInstallRunArgs(t *testing.T, name string, args []string) {
	t.Helper()
	if name != "bash" {
		t.Errorf("run name = %q, want bash", name)
	}
	if len(args) != 2 || args[0] != "-c" {
		t.Errorf("run args = %v, want [-c, <command>]", args)
		return
	}
	if !strings.Contains(args[1], "NONINTERACTIVE=1") {
		t.Errorf("run command %q missing NONINTERACTIVE=1", args[1])
	}
	if !strings.Contains(args[1], "/tmp/macheim-fake.sh") {
		t.Errorf("run command %q missing temp script path", args[1])
	}
}

func TestInstall_FreshInstall(t *testing.T) {
	t.Parallel()
	spy := &freshInstallSpy{}
	s := newFreshInstallSeam(t, spy)
	if err := install(s, &config.Runtime{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	assertFreshInstallSpy(t, spy)
}

func TestInstall_UnsupportedArch(t *testing.T) {
	t.Parallel()
	s := stubInstallSeam(installSeam{
		detect: func() (string, bool) { return "", false },
	})
	err := install(s, &config.Runtime{})
	if err == nil {
		t.Fatal("err = nil, want error for unsupported arch")
	}
	if !strings.Contains(err.Error(), "unsupported architecture") {
		t.Errorf("err = %q, want substring 'unsupported architecture'", err.Error())
	}
}

func TestInstall_DryRunSkipsSideEffects(t *testing.T) {
	t.Parallel()
	var runCalled bool
	s := stubInstallSeam(installSeam{
		// Pretend brew is not installed so we exercise the install path.
		detect: func() (string, bool) { return "/opt/homebrew/bin/brew", false },
		scripts: func() ([]byte, error) {
			return []byte("#!/bin/bash\n"), nil
		},
		tempFile: func([]byte) (string, func(), error) {
			return "/tmp/macheim-dry.sh", func() {}, nil
		},
		run: func(rt *config.Runtime, _ string, _ ...string) (string, error) {
			runCalled = true
			// In a real DryRun the run function would short-circuit
			// before exec — confirm we received the dry-run runtime.
			if rt == nil || !rt.DryRun {
				t.Error("run invoked with non-dry-run runtime; want dry-run propagated")
			}
			return "", nil
		},
		appendIfMissing: func(rt *config.Runtime, _, _ string) error {
			if rt == nil || !rt.DryRun {
				t.Error("appendIfMissing invoked with non-dry-run runtime; want dry-run propagated")
			}
			return nil
		},
		version: func(string) (string, error) {
			t.Error("version called under dry-run; want skip")
			return "", nil
		},
	})

	err := install(s, &config.Runtime{DryRun: true})
	if err != nil {
		t.Fatalf("install: %v", err)
	}
	if !runCalled {
		t.Error("shell.Run never invoked; dry-run should still drive Run (which itself prints [dry-run])")
	}
}

func TestInstall_PostInstallDetectFails(t *testing.T) {
	t.Parallel()
	s := stubInstallSeam(installSeam{
		// Always returns not-installed, even after the installer runs.
		detect:  func() (string, bool) { return "/opt/homebrew/bin/brew", false },
		scripts: func() ([]byte, error) { return []byte("#!/bin/bash\n"), nil },
		tempFile: func([]byte) (string, func(), error) {
			return "/tmp/macheim-fake.sh", func() {}, nil
		},
		run: func(*config.Runtime, string, ...string) (string, error) { return "", nil },
	})
	err := install(s, &config.Runtime{})
	if err == nil {
		t.Fatal("err = nil, want error when post-install detect fails")
	}
	if !strings.Contains(err.Error(), "installer ran but brew not found") {
		t.Errorf("err = %q, want 'installer ran but brew not found'", err.Error())
	}
}

func TestInstall_ScriptsReadFails(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("embedded read failure")
	s := stubInstallSeam(installSeam{
		detect:  func() (string, bool) { return "/opt/homebrew/bin/brew", false },
		scripts: func() ([]byte, error) { return nil, wantErr },
	})
	err := install(s, &config.Runtime{})
	if err == nil {
		t.Fatal("err = nil, want error when embedded read fails")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want errors.Is wrapping %v", err, wantErr)
	}
}

func TestInstall_UnknownShellWarnsAndContinues(t *testing.T) {
	t.Parallel()
	var (
		appendCalled  bool
		versionCalled bool
	)
	s := stubInstallSeam(installSeam{
		detect:      func() (string, bool) { return "/opt/homebrew/bin/brew", true },
		shellDetect: func() (string, string, error) { return "", "", errors.New("unknown shell") },
		appendIfMissing: func(*config.Runtime, string, string) error {
			appendCalled = true
			return nil
		},
		version: func(string) (string, error) {
			versionCalled = true
			return "4.3.5", nil
		},
	})
	if err := install(s, &config.Runtime{}); err != nil {
		t.Fatalf("install: %v", err)
	}
	if appendCalled {
		t.Error("appendIfMissing called despite unknown shell; want skip")
	}
	if !versionCalled {
		t.Error("version verification not called; install should still verify post-install")
	}
}

func TestInstall_VersionFailureIsInformational(t *testing.T) {
	t.Parallel()
	s := stubInstallSeam(installSeam{
		detect:  func() (string, bool) { return "/opt/homebrew/bin/brew", true },
		version: func(string) (string, error) { return "", errors.New("brew --version failed") },
	})
	if err := install(s, &config.Runtime{}); err != nil {
		t.Errorf("install returned error on version failure: %v; want nil (version failure is informational)", err)
	}
}
