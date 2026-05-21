package doctor

import (
	"bytes"
	"strings"
	"testing"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/output"
)

// newTestRender constructs a render bound to a bytes.Buffer. Because the
// buffer is not an *os.File, shouldUseColor returns false regardless of
// rt.NoColor — making test output deterministic.
func newTestRender(rt *config.Runtime) (*render, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return newRender(rt, buf), buf
}

func TestRender_PassRow_Default(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{})
	r.row("xcode-select", Result{OK: true, Probe: "xcode-select -p → /x"})
	got := buf.String()
	if got != "✓ xcode-select\n" {
		t.Errorf("got %q, want %q", got, "✓ xcode-select\n")
	}
}

func TestRender_PassRow_VerboseShowsProbe(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{Verbose: true})
	r.row("xcode-select", Result{OK: true, Probe: "xcode-select -p → /x"})
	got := buf.String()
	want := "✓ xcode-select\n   probed: xcode-select -p → /x\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRender_PassRow_QuietSuppresses(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{Quiet: true})
	r.row("xcode-select", Result{OK: true, Probe: "x"})
	if buf.Len() != 0 {
		t.Errorf("quiet should suppress pass rows, got %q", buf.String())
	}
}

func TestRender_FailRow_IncludesRemediation(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{})
	r.row("brew", Result{OK: false, Probe: "/brew (not found)", Remediation: "Run: macheim brew install"})
	got := buf.String()
	want := "✗ brew\n   → Run: macheim brew install\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRender_FailRow_QuietStillPrints(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{Quiet: true})
	r.row("brew", Result{OK: false, Remediation: "Run: macheim brew install"})
	if !strings.Contains(buf.String(), "✗ brew") {
		t.Errorf("quiet must still print failures, got %q", buf.String())
	}
}

func TestRender_FailRow_VerboseShowsProbeAboveRemediation(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{Verbose: true})
	r.row("brew", Result{OK: false, Probe: "/opt/homebrew/bin/brew (not found)", Remediation: "Run: macheim brew install"})
	got := buf.String()
	want := "✗ brew\n   probed: /opt/homebrew/bin/brew (not found)\n   → Run: macheim brew install\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRender_Summary_AllPass(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{})
	r.summary(0)
	if buf.String() != "All checks passed.\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestRender_Summary_SingularPlural(t *testing.T) {
	t.Parallel()
	for failed, want := range map[int]string{
		1: "1 check failed.\n",
		3: "3 checks failed.\n",
	} {
		r, buf := newTestRender(&config.Runtime{})
		r.summary(failed)
		if buf.String() != want {
			t.Errorf("failed=%d: got %q, want %q", failed, buf.String(), want)
		}
	}
}

// TestRender_ColorRespected ensures the ANSI escapes ARE emitted when
// useColor is forced on. The non-TTY default in newTestRender suppresses
// them; this test pokes the field directly.
func TestRender_ColorRespected(t *testing.T) {
	t.Parallel()
	r, buf := newTestRender(&config.Runtime{})
	r.useColor = true
	r.row("ok", Result{OK: true})
	r.row("bad", Result{OK: false, Remediation: "fix"})
	r.summary(1)
	got := buf.String()
	for _, want := range []string{output.AnsiGreen, output.AnsiRed, output.AnsiReset} {
		if !strings.Contains(got, want) {
			t.Errorf("output should contain ANSI %q; got %q", want, got)
		}
	}
}

func TestRender_NoColor_StripsANSI(t *testing.T) {
	t.Parallel()
	// NoColor + non-TTY (bytes.Buffer) both push useColor to false.
	r, buf := newTestRender(&config.Runtime{NoColor: true})
	r.row("bad", Result{OK: false, Remediation: "fix"})
	r.summary(1)
	got := buf.String()
	for _, banned := range []string{output.AnsiGreen, output.AnsiRed, output.AnsiReset} {
		if strings.Contains(got, banned) {
			t.Errorf("output should not contain ANSI %q; got %q", banned, got)
		}
	}
}
