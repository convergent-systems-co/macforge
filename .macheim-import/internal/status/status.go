// Package status implements `macheim status`: a read-only summary of
// what macheim sees right now. Always exits 0; sections may report
// "?" markers for not-yet-implemented checks.
package status

import (
	"io"

	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/output"
)

// Section is one named area of the status report.
type Section struct {
	Name string
	Run  func(rt *config.Runtime) []Row
}

// Row is one renderable line in a section's output.
type Row struct {
	Marker  output.Marker
	Name    string
	Detail  string
	Verbose string
	// Hidden rows are omitted when rt.Quiet is true. Used for "?" placeholder
	// rows that report deferred work — noise for an operator wanting actionable
	// state.
	Hidden bool
}

// DefaultSections returns the canonical list of sections, in display order.
func DefaultSections() []Section {
	return []Section{
		brewSection(),
		repoSection(),
		driftSection(),
	}
}

// Run executes every section in order and prints the rows to w. Always
// returns nil — status is informational and never fails.
func Run(rt *config.Runtime, w io.Writer) error {
	useColor := output.UseColor(rt.NoColor, w)
	for _, s := range DefaultSections() {
		for _, row := range s.Run(rt) {
			if rt.Quiet && row.Hidden {
				continue
			}
			output.Row(w, useColor, rt.Verbose, row.Marker, row.Name, row.Detail, row.Verbose)
		}
	}
	return nil
}
