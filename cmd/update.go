// Package cmd's update subtree houses the two-way sync between this Mac
// and the repo. Per GOALS.md, both subcommands accept --module to scope
// the operation; remote-to-local additionally accepts --prune (drop brew
// formulae no longer in the repo Brewfile) and --no-pull (skip the git
// pull). The global --yes and --dry-run flags live on the root command
// and reach the module implementations via rt.
//
// Dispatch is intentionally thin: this file parses flags and routes to
// per-module entry points in internal/brew and internal/dotfiles. The
// "defaults" module is a stub in both directions until its own sub-epic
// lands.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/polliard/macheim/internal/brew"
	"github.com/polliard/macheim/internal/config"
	"github.com/polliard/macheim/internal/dotfiles"
	"github.com/urfave/cli/v3"
)

// moduleAll is the catch-all module value that expands to every
// implemented module. It is the default for --module.
const moduleAll = "all"

// updateCommand wires the `macheim update` subtree. Both subcommands
// share the --module flag; remote-to-local adds --prune and --no-pull.
// --yes and --dry-run are inherited from the root command and read off
// rt by the module implementations.
func updateCommand(rt *config.Runtime) *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Reconcile state between this Mac and the repo",
		Commands: []*cli.Command{
			{
				Name:  "local-to-remote",
				Usage: "Update the repo to match local state",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "module",
						Value: moduleAll,
						Usage: "brew|dotfiles|defaults|all",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return updateLocalToRemote(rt, cmd.String("module"))
				},
			},
			{
				Name:  "remote-to-local",
				Usage: "Update this Mac to match the repo",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "module",
						Value: moduleAll,
						Usage: "brew|dotfiles|defaults|all",
					},
					&cli.BoolFlag{
						Name:  "prune",
						Usage: "Remove brew formulae no longer in the Brewfile",
					},
					&cli.BoolFlag{
						Name:  "no-pull",
						Usage: "Skip git pull before applying",
					},
				},
				Action: func(_ context.Context, cmd *cli.Command) error {
					return updateRemoteToLocal(rt, cmd.String("module"), cmd.Bool("prune"), cmd.Bool("no-pull"))
				},
			},
		},
	}
}

// dispatchModules expands the --module flag to the ordered list of
// modules the caller wants to act on. "all" expands to brew, dotfiles,
// defaults (in that order — brew touches the most state and surfaces
// drift earliest); any other value is returned as a single-element
// slice. Unknown values are validated by the per-direction dispatcher
// before any module runs.
func dispatchModules(module string) []string {
	if module == moduleAll {
		return []string{"brew", "dotfiles", "defaults"}
	}
	return []string{module}
}

// validateModule rejects values that are neither a known module name nor
// "all". The check runs before any module action so a typo never causes
// partial work.
func validateModule(module string) error {
	switch module {
	case moduleAll, "brew", "dotfiles", "defaults":
		return nil
	default:
		return fmt.Errorf("update: unknown module %q (want brew, dotfiles, defaults, or all)", module)
	}
}

// updateLocalToRemote dispatches each requested module's local-to-remote
// handler in order. On success it prints the canonical "review with git
// diff" footer to stderr — that footer is part of the update contract
// per GOALS.md so the user knows the next manual step.
func updateLocalToRemote(rt *config.Runtime, module string) error {
	if err := validateModule(module); err != nil {
		return err
	}
	for _, m := range dispatchModules(module) {
		if err := dispatchLocalToRemote(rt, m); err != nil {
			return err
		}
	}
	_, _ = fmt.Fprintln(os.Stderr, "Repo updated. Review with `git diff`, then commit and rebuild with `make build` to refresh the embedded fallback.")
	return nil
}

// dispatchLocalToRemote routes one module to its package-level handler.
// Returns errNotImplemented for "defaults" — a stub the user sees as
// "defaults: not implemented yet" on stderr without aborting the rest
// of the run.
func dispatchLocalToRemote(rt *config.Runtime, module string) error {
	switch module {
	case "brew":
		if err := brew.UpdateLocalToRemote(rt); err != nil {
			return fmt.Errorf("brew local-to-remote: %w", err)
		}
	case "dotfiles":
		if err := dotfiles.UpdateLocalToRemote(rt); err != nil {
			return fmt.Errorf("dotfiles local-to-remote: %w", err)
		}
	case "defaults":
		_, _ = fmt.Fprintln(os.Stderr, "defaults: not implemented yet")
	default:
		// validateModule guards this path, but the type switch is exhaustive
		// for the reader.
		return errors.New("update local-to-remote: unreachable module dispatch")
	}
	return nil
}

// updateRemoteToLocal dispatches each requested module's remote-to-local
// handler with the prune / no-pull flags threaded through. Unlike
// local-to-remote there is no closing footer — the user already sees
// per-module output as the changes apply.
func updateRemoteToLocal(rt *config.Runtime, module string, prune, noPull bool) error {
	if err := validateModule(module); err != nil {
		return err
	}
	for _, m := range dispatchModules(module) {
		if err := dispatchRemoteToLocal(rt, m, prune, noPull); err != nil {
			return err
		}
	}
	return nil
}

// dispatchRemoteToLocal routes one module to its package-level handler.
// Like dispatchLocalToRemote, "defaults" is a stub.
func dispatchRemoteToLocal(rt *config.Runtime, module string, prune, noPull bool) error {
	switch module {
	case "brew":
		if err := brew.UpdateRemoteToLocal(rt, prune, noPull); err != nil {
			return fmt.Errorf("brew remote-to-local: %w", err)
		}
	case "dotfiles":
		if err := dotfilesRemoteToLocal(rt, noPull); err != nil {
			return fmt.Errorf("dotfiles remote-to-local: %w", err)
		}
	case "defaults":
		_, _ = fmt.Fprintln(os.Stderr, "defaults: not implemented yet")
	default:
		return errors.New("update remote-to-local: unreachable module dispatch")
	}
	return nil
}

// dotfilesRemoteToLocal applies the repo's dotfiles tree to $HOME. It
// resolves the repo path, fails fast in embed-fallback mode (dotfiles
// apply has no embedded source today), and forwards to dotfiles.Apply
// which already implements the forward direction with backups.
//
// noPull is accepted for symmetry with brew but currently unused —
// dotfiles read directly from the repo working tree, so the pull (if
// any) happens at the outer dispatcher level when brew also runs. A
// future change can introduce a per-module pull guard here without
// changing the public flag set.
func dotfilesRemoteToLocal(rt *config.Runtime, _ bool) error {
	repoPath, _, err := rt.ResolveRepoPath()
	if err != nil {
		return err
	}
	if repoPath == "" {
		return errors.New("update remote-to-local: no repo configured; clone the macheim repo and set --repo or MACHEIM_REPO first")
	}
	homePath, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	_, err = dotfiles.Apply(rt, repoPath, homePath)
	return err
}
