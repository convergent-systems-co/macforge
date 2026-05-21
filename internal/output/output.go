// Package output provides the shared row renderer used by `macheim doctor`
// and `macheim status` (and any future read-only command that needs the
// same visual language). Intentionally has no dependency on internal/config
// or the urfave framework — pure rendering primitives.
package output

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// ANSI escape sequences used when color is enabled. Exported so consumers
// (and their tests) can reference the same constants the renderer emits.
const (
	AnsiGreen  = "\033[32m"
	AnsiRed    = "\033[31m"
	AnsiYellow = "\033[33m"
	AnsiReset  = "\033[0m"
)

// Marker is the glyph rendered at the head of a Row.
type Marker int

// Marker values for Row's leading glyph. MarkerOK renders as a green "✓"
// (affirmative present / pass), MarkerFail as a red "✗" (confirmed absent
// or failed), MarkerUnknown as a yellow "?" (unknown / not implemented /
// deferred).
const (
	MarkerOK Marker = iota
	MarkerFail
	MarkerUnknown
)

// UseColor reports whether ANSI escapes should be emitted. Returns false
// when the caller requests no color, when w is not an *os.File (e.g. a
// bytes.Buffer in tests, or an io.Pipe), or when the file is not a TTY.
func UseColor(noColor bool, w io.Writer) bool {
	if noColor {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// Row prints one labeled line:
//
//	<marker> <name>            (when detail is empty)
//	<marker> <name>  <detail>  (when detail is non-empty)
//
// When verbose is true AND probe is non-empty, an indented context line
// follows. The marker is colorized iff useColor.
func Row(w io.Writer, useColor, verbose bool, m Marker, name, detail, probe string) {
	glyph, code := markerGlyph(m)
	if useColor {
		_, _ = fmt.Fprintf(w, "%s%s%s %s", code, glyph, AnsiReset, name)
	} else {
		_, _ = fmt.Fprintf(w, "%s %s", glyph, name)
	}
	if detail != "" {
		_, _ = fmt.Fprintf(w, "  %s", detail)
	}
	_, _ = fmt.Fprintln(w)
	if verbose && probe != "" {
		_, _ = fmt.Fprintf(w, "   %s\n", probe)
	}
}

// Summary prints a single colored line, used by callers that emit a tally
// after a series of Rows (e.g. doctor's "All checks passed." trailer).
// Empty text is a no-op.
func Summary(w io.Writer, useColor bool, m Marker, text string) {
	if text == "" {
		return
	}
	if !useColor {
		_, _ = fmt.Fprintln(w, text)
		return
	}
	_, code := markerGlyph(m)
	_, _ = fmt.Fprintf(w, "%s%s%s\n", code, text, AnsiReset)
}

// markerGlyph maps a Marker to its (glyph, ansi-code) pair. The reset
// suffix is the caller's responsibility.
func markerGlyph(m Marker) (string, string) {
	switch m {
	case MarkerOK:
		return "✓", AnsiGreen
	case MarkerFail:
		return "✗", AnsiRed
	case MarkerUnknown:
		return "?", AnsiYellow
	default:
		return "?", AnsiReset
	}
}
