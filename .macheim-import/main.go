// Package main is the entry point for the macheim CLI.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/polliard/macheim/cmd"
	"github.com/polliard/macheim/internal/config"
)

// Build-time identity, populated via -ldflags from the Makefile.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	rt := &config.Runtime{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
	}
	root := cmd.NewRoot(rt)
	if err := root.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "macheim:", err)
		os.Exit(1)
	}
}
