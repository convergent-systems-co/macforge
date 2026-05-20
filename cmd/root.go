// Package cmd holds the CLI command tree built on urfave/cli/v3.
package cmd

import (
	"context"
	"fmt"

	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

// init drops the "-v" alias from urfave/cli/v3's built-in --version flag.
// GOALS.md gives --verbose the "-v" short alias; the framework's default
// --version flag also claims "-v", and the parser collision silently routes
// "--verbose" into version-print mode. Rebinding here once (idempotent)
// frees "-v" for --verbose. --version long-form still works.
func init() {
	if vf, ok := cli.VersionFlag.(*cli.BoolFlag); ok {
		vf.Aliases = nil
	}
}

// notImplemented is the stub body every unimplemented subcommand uses. It
// prints the canonical "see issue #N" line and exits 0. The name argument
// is the fully-qualified command path (e.g. "brew bundle"); issue is the
// GitHub issue number that owns the real implementation.
//
// Writes to cmd.Root().Writer (not cmd.Writer) because urfave/cli/v3 defaults
// a subcommand's Writer to os.Stdout independently of the root, so test
// capture via root.Writer would otherwise be invisible from a subcommand.
func notImplemented(name string, issue int) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		_, _ = fmt.Fprintf(cmd.Root().Writer, "%s: not implemented yet (see issue #%d)\n", name, issue)
		return nil
	}
}

// NewRoot returns the root *cli.Command. The caller passes a *config.Runtime;
// the Before hook populates it from parsed flags. Subcommands hold the same
// pointer and read populated values from their Action handlers.
func NewRoot(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:                  "macheim",
		Usage:                 "Bootstrap and sync a Mac to a known-good state defined in a git repo.",
		Version:               rt.VersionString(),
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "repo",
				Usage:   "Path to the macheim repo (overrides discovery)",
				Sources: cli.EnvVars("MACHEIM_REPO"),
			},
			&cli.BoolFlag{Name: "dry-run", Usage: "Print actions, change nothing"},
			&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Extra output"},
			&cli.BoolFlag{Name: "quiet", Aliases: []string{"q"}, Usage: "Suppress non-error output"},
			&cli.BoolFlag{Name: "yes", Aliases: []string{"y"}, Usage: "Skip confirmation prompts"},
			&cli.BoolFlag{Name: "no-color", Usage: "Disable colored output"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			rt.RepoPath = cmd.String("repo")
			rt.DryRun = cmd.Bool("dry-run")
			rt.Verbose = cmd.Bool("verbose")
			rt.Quiet = cmd.Bool("quiet")
			rt.Yes = cmd.Bool("yes")
			rt.NoColor = cmd.Bool("no-color")
			if err := rt.Validate(); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			bootstrapCommand(rt),
			brewCommand(rt),
			zshCommand(rt),
			dotfilesCommand(rt),
			macosCommand(rt),
			downloadsCommand(rt),
			updateCommand(rt),
			statusCommand(rt),
			doctorCommand(rt),
		},
	}
}
