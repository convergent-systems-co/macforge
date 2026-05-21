package doctor

import (
	"fmt"
	"io"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/output"
)

// render is the per-Run rendering state. One instance is constructed when
// Run begins and used for every row plus the summary line. Holds the
// per-Run booleans derived from Runtime so they're computed once.
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
		useColor: output.UseColor(rt.NoColor, w),
	}
}

// row prints one check's outcome via internal/output, plus the doctor-
// specific "→ remediation" trailing line on failure. Pass rows are
// suppressed under --quiet; fail rows always print.
func (r *render) row(name string, res Result) {
	probe := ""
	if res.Probe != "" {
		probe = "probed: " + res.Probe
	}
	if res.OK {
		if r.quiet {
			return
		}
		output.Row(r.w, r.useColor, r.verbose, output.MarkerOK, name, "", probe)
		return
	}
	output.Row(r.w, r.useColor, r.verbose, output.MarkerFail, name, "", probe)
	if res.Remediation != "" {
		_, _ = fmt.Fprintf(r.w, "   → %s\n", res.Remediation)
	}
}

// summary prints the final line via internal/output. Always emitted
// regardless of --quiet.
func (r *render) summary(failed int) {
	if failed == 0 {
		output.Summary(r.w, r.useColor, output.MarkerOK, "All checks passed.")
		return
	}
	noun := "checks"
	if failed == 1 {
		noun = "check"
	}
	output.Summary(r.w, r.useColor, output.MarkerFail, fmt.Sprintf("%d %s failed.", failed, noun))
}
