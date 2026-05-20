package doctor

import (
	"fmt"
	"io"
	"os"

	"github.com/polliard/macheim/internal/config"
	"golang.org/x/term"
)

const (
	ansiGreen = "\033[32m"
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

// render is the per-Run rendering state. One instance is constructed when
// Run begins and used for every row plus the summary line.
type render struct {
	w        io.Writer
	quiet    bool
	verbose  bool
	useColor bool
}

func newRender(rt *config.Runtime, w io.Writer) *render {
	return &render{
		w:        w,
		quiet:    rt.Quiet,
		verbose:  rt.Verbose,
		useColor: shouldUseColor(rt, w),
	}
}

// shouldUseColor returns true when ANSI escapes should be emitted: never
// when rt.NoColor is set, and only when the writer is a real terminal.
// When w is not an *os.File (test bytes.Buffer, pipe, etc.) treat as
// non-TTY so test output stays deterministic.
func shouldUseColor(rt *config.Runtime, w io.Writer) bool {
	if rt.NoColor {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func (r *render) colorize(code, s string) string {
	if !r.useColor {
		return s
	}
	return code + s + ansiReset
}

func (r *render) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(r.w, format, args...)
}

// row prints one check's outcome. Pass rows are suppressed under --quiet;
// fail rows always print. --verbose adds an indented probe line.
func (r *render) row(name string, res Result) {
	if res.OK {
		if r.quiet {
			return
		}
		r.printf("%s %s\n", r.colorize(ansiGreen, "✓"), name)
		if r.verbose && res.Probe != "" {
			r.printf("   probed: %s\n", res.Probe)
		}
		return
	}
	r.printf("%s %s\n", r.colorize(ansiRed, "✗"), name)
	if r.verbose && res.Probe != "" {
		r.printf("   probed: %s\n", res.Probe)
	}
	if res.Remediation != "" {
		r.printf("   → %s\n", res.Remediation)
	}
}

// summary prints the final line. Always emitted regardless of --quiet.
func (r *render) summary(failed int) {
	if failed == 0 {
		r.printf("%s\n", r.colorize(ansiGreen, "All checks passed."))
		return
	}
	noun := "checks"
	if failed == 1 {
		noun = "check"
	}
	r.printf("%s\n", r.colorize(ansiRed, fmt.Sprintf("%d %s failed.", failed, noun)))
}
