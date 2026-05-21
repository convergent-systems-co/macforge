//go:build !windows

package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestRow_PlainOK(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, false, MarkerOK, "name", "", "")
	if got := buf.String(); got != "✓ name\n" {
		t.Errorf("got %q, want %q", got, "✓ name\n")
	}
}

func TestRow_PlainFail(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, false, MarkerFail, "name", "", "")
	if got := buf.String(); got != "✗ name\n" {
		t.Errorf("got %q, want %q", got, "✗ name\n")
	}
}

func TestRow_PlainUnknown(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, false, MarkerUnknown, "name", "", "")
	if got := buf.String(); got != "? name\n" {
		t.Errorf("got %q, want %q", got, "? name\n")
	}
}

func TestRow_WithDetail(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, false, MarkerOK, "brew", "/opt/homebrew/bin/brew (4.3.5)", "")
	want := "✓ brew  /opt/homebrew/bin/brew (4.3.5)\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRow_VerboseAddsProbe(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, true, MarkerOK, "xcode-select", "", "probed: xcode-select -p → /x")
	want := "✓ xcode-select\n   probed: xcode-select -p → /x\n"
	if got := buf.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRow_VerboseFalseSuppressesProbe(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, false, MarkerOK, "name", "", "probed: something")
	if got := buf.String(); got != "✓ name\n" {
		t.Errorf("expected probe hidden, got %q", got)
	}
}

func TestRow_VerboseEmptyProbeNoIndentLine(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, true, MarkerOK, "name", "", "")
	if got := buf.String(); got != "✓ name\n" {
		t.Errorf("empty probe should not produce indent line, got %q", got)
	}
}

func TestRow_ColorEmitsANSI(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, true, false, MarkerOK, "name", "", "")
	got := buf.String()
	for _, want := range []string{AnsiGreen, AnsiReset, "✓", "name"} {
		if !strings.Contains(got, want) {
			t.Errorf("colored output should contain %q; got %q", want, got)
		}
	}
}

func TestRow_NoColorStripsANSI(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Row(buf, false, false, MarkerOK, "name", "", "")
	got := buf.String()
	for _, banned := range []string{AnsiGreen, AnsiRed, AnsiYellow, AnsiReset} {
		if strings.Contains(got, banned) {
			t.Errorf("non-color output must not contain %q; got %q", banned, got)
		}
	}
}

func TestSummary_PlainAllPass(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Summary(buf, false, MarkerOK, "All checks passed.")
	if got := buf.String(); got != "All checks passed.\n" {
		t.Errorf("got %q", got)
	}
}

func TestSummary_Empty(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Summary(buf, true, MarkerOK, "")
	if buf.Len() != 0 {
		t.Errorf("empty summary should produce no output, got %q", buf.String())
	}
}

func TestSummary_Colorized(t *testing.T) {
	t.Parallel()
	buf := &bytes.Buffer{}
	Summary(buf, true, MarkerFail, "1 check failed.")
	got := buf.String()
	for _, want := range []string{AnsiRed, AnsiReset, "1 check failed."} {
		if !strings.Contains(got, want) {
			t.Errorf("colored summary should contain %q; got %q", want, got)
		}
	}
}

func TestUseColor_NoColorFlagBeatsTTY(t *testing.T) {
	t.Parallel()
	if UseColor(true, &bytes.Buffer{}) {
		t.Error("noColor=true must always disable color")
	}
}

func TestUseColor_BytesBufferAlwaysFalse(t *testing.T) {
	t.Parallel()
	if UseColor(false, &bytes.Buffer{}) {
		t.Error("bytes.Buffer is not an *os.File; UseColor must return false")
	}
}