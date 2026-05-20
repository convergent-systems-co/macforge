// Package main is the entry point for the macheim CLI.
package main

// Build-time identity, populated via -ldflags. Sub-issue #5 wires these into
// the urfave/cli/v3 root command's --version flag.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// Command tree lands in sub-issue #5.
	_ = version
	_ = commit
	_ = buildDate
}
