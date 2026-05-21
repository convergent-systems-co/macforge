# Design — CLI shell (sub-epic #5)

**Tracks:** [Index #3](https://github.com/polliard/macheim/issues/3) → [Epic #25 Foundation](https://github.com/polliard/macheim/issues/25) → sub-epic [#5](https://github.com/polliard/macheim/issues/5)
**Status:** Draft — awaiting user review
**Date:** 2026-05-20
**Author:** Thomas Polliard (with Claude Opus 4.7)

This is PR A of a three-PR chunk:

| PR | Sub-epic | Scope |
|---|---|---|
| **A (this one)** | #5 | CLI command tree, global `--dry-run` flag, every subcommand registered as a "not implemented yet" stub |
| B | #10 | Working `doctor` (xcode-select, brew presence + arch, repo path, write permissions, shell rc detection) |
| C | #8 + #47 | `internal/shell/run.go` live-streaming exec helper that honors `--dry-run`/`--quiet`/`--verbose`; migrate doctor onto it |

---

## 1. Objective

Establish the user-visible CLI surface. After this PR, `macheim --help` lists every command from `GOALS.md`; each subcommand exists and prints `"<name>: not implemented yet (see issue #NN)"` and exits 0; every command inherits the six global flags (`--repo`, `--dry-run`, `--verbose`, `--quiet`, `--yes`, `--no-color`); and `--version` reports the ldflags-injected build identity.

No subcommand does real work yet. The point is to land the surface once so later sub-epics are pure feature work — never feature-plus-flag-plumbing.

## 2. Rationale

`GOALS.md` and issue #5 already specify the command tree and flag set. The remaining design choices are:

### Alternatives table — flag propagation pattern

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Package-level `var rt config.Runtime` singleton | Simplest access from any subcommand | Globals are a test-isolation hazard and a concurrency hazard if anything ever runs in parallel | Rejected |
| `context.WithValue` carrying `*Runtime` | Idiomatic Go context plumbing | urfave/cli/v3's Before returns `(context.Context, error)`; mixing typed context values with `*cli.Command.String()` access is two ways to read the same data | Rejected |
| Pointer-passing: every subcommand constructor takes `*Runtime`, Before writes through the pointer | One ownership path; trivially testable (pass a fresh `*Runtime` in tests); no globals | Constructor signatures all gain a parameter | **Chosen** |

### Alternatives table — file layout for stubs

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| One `cmd/stubs.go` collecting every unimplemented stub | Minimises file count for sub-epic #5 | Every later sub-issue that implements one of those commands has to *move* code out of `stubs.go` (more churn, noisier diffs) | Rejected |
| One file per subcommand from day one | Matches the long-term layout `GOALS.md` describes; later sub-issues swap stub-body for real-body in-place | More files for this PR (each ~10 lines) | **Chosen** |

### Alternatives table — `--no-color` enforcement

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Build the output helper package now | One-shot; everything respects the flag from day one | No ANSI exists to suppress; speculative infra | Rejected |
| Store the flag in `Runtime` only; document that future output helpers MUST consult `rt.NoColor` | YAGNI; flag's AC is "is settable and stored," which is met | Risk: a later contributor adds `fmt.Println("\033[...]")` without consulting `Runtime` | **Chosen** |

Mitigation for the chosen path: a `cmd/CLAUDE.md`-equivalent note in `Runtime`'s doc-comment and a lint-grade rule in the followup output-helper sub-issue.

### Alternatives table — `--verbose` vs `--quiet` interaction

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Both can be set; quiet wins | Permissive | Confusing — user wrote `-v -q` and got `-q` behavior silently | Rejected |
| Both can be set; return error from Before | Explicit | One-line guard, clearest signal | **Chosen** |

## 3. Scope

### Files to create

- `internal/config/runtime.go` — `Runtime` struct + `Validate()` (mutex check on `--verbose`/`--quiet`)
- `internal/config/runtime_test.go` — table-driven tests
- `cmd/root.go` — `NewRoot(rt *config.Runtime) *cli.Command` with global flags + Before hook
- `cmd/bootstrap.go` — stub
- `cmd/brew.go` — group: `brew install` + `brew bundle` stubs
- `cmd/doctor.go` — stub (real work lands in PR B / issue #10)
- `cmd/downloads.go` — stub
- `cmd/dotfiles.go` — stub: `dotfiles apply`
- `cmd/macos.go` — stub: `macos defaults`
- `cmd/status.go` — stub
- `cmd/update.go` — group: `update local-to-remote` + `update remote-to-local` stubs
- `cmd/zsh.go` — stub: `zsh setup`
- `cmd/root_test.go` — integration: --help shows every command, --version, stubs exit 0, Before validation

### Files to modify

- `main.go` — wire `cmd.NewRoot`, pass `version`/`commit`/`buildDate` into Runtime, call `cmd.Run(ctx, os.Args)`
- Remove the three `_ = X` placeholder lines (they become real consumers)

### Files NOT touched

- `cmd/.gitkeep` — deleted by virtue of new real files in `cmd/`
- Everything under `internal/embedded/`, `internal/shell/`, `internal/brew/`, `internal/dotfiles/`, `internal/gitrepo/` — those are later sub-issues
- `tools.go` — its blank imports are now real imports from `cmd/`, so `go mod tidy` may remove some lines. Re-evaluate at end of PR.

## 4. `internal/config/runtime.go`

```go
// Package config holds the runtime configuration parsed from CLI flags and
// (later) from ~/.config/macheim/config.yaml. Subcommands receive a *Runtime
// pointer at construction time; the root command's Before hook populates it
// from flag values before any Action runs.
package config

import (
	"errors"
	"fmt"
)

// Runtime captures the inherited global flags and the build-time identity.
// All fields are zero-valued on construction; the root command's Before hook
// writes them.
//
// Output helpers added in a later sub-issue MUST consult NoColor before
// emitting ANSI. Until that helper lands no command emits ANSI, so the flag
// is stored-but-unused — which satisfies issue #5's "--no-color suppresses
// all ANSI globally" AC trivially.
type Runtime struct {
	// Inherited global flags
	RepoPath string
	DryRun   bool
	Verbose  bool
	Quiet    bool
	Yes      bool
	NoColor  bool

	// Build-time identity, set in main() from -ldflags-injected vars.
	Version   string
	Commit    string
	BuildDate string
}

// Validate enforces flag-interaction rules. Called from the root command's
// Before hook after the flags have been parsed.
func (r *Runtime) Validate() error {
	if r.Verbose && r.Quiet {
		return errors.New("--verbose and --quiet are mutually exclusive")
	}
	return nil
}

// VersionString formats the build identity for --version output.
func (r *Runtime) VersionString() string {
	return fmt.Sprintf("%s (commit %s, built %s)", r.Version, r.Commit, r.BuildDate)
}
```

### `runtime_test.go` shape

Table-driven against `Validate`:

| Case | Verbose | Quiet | Expected |
|---|---|---|---|
| both off | false | false | nil |
| only verbose | true | false | nil |
| only quiet | false | true | nil |
| both on | true | true | error |

Plus a `VersionString` golden-string test with fixed inputs.

## 5. `cmd/root.go`

```go
// Package cmd holds the CLI command tree built on urfave/cli/v3.
package cmd

import (
	"context"
	"fmt"

	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

// NewRoot returns the root *cli.Command with every subcommand registered.
// The caller passes a *config.Runtime; the Before hook populates it from
// parsed flags. Subcommands hold the same pointer and read populated values
// in their Action handlers.
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

// notImplemented is the stub body every unimplemented subcommand uses. It
// prints the canonical "see issue #N" line, ignores rt for now, and exits 0.
// Each stub call-site passes the unique GitHub issue number that implements it.
func notImplemented(name string, issue int) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		fmt.Fprintf(cmd.Writer, "%s: not implemented yet (see issue #%d)\n", name, issue)
		return nil
	}
}
```

## 6. Stub files

Each unimplemented command lives in its own file, matching the long-term layout. Sample:

```go
// cmd/doctor.go
package cmd

import (
	"github.com/polliard/macheim/internal/config"
	"github.com/urfave/cli/v3"
)

func doctorCommand(rt *config.Runtime) *cli.Command {
	_ = rt // populated by root's Before; consumed in issue #10
	return &cli.Command{
		Name:   "doctor",
		Usage:  "Sanity-check the macheim environment",
		Action: notImplemented("doctor", 10),
	}
}
```

Issue-number mapping for the `notImplemented` calls (verified against the GitHub issue list on 2026-05-20):

| Command | Issue |
|---|---|
| `bootstrap` | #19 (`[Orchestration] bootstrap: end-to-end runner`) |
| `brew install` | #12 |
| `brew bundle` | #13 |
| `zsh setup` | #20 (combined stubs) |
| `dotfiles apply` | #17 (`[Dotfiles] Dotfiles diff + remote-to-local apply` — "apply" is the remote-to-local direction) |
| `macos defaults` | #20 |
| `downloads` | #20 |
| `update local-to-remote` | #15 (`[Brew] update local-to-remote --module=brew` — first module to implement; dotfiles direction is #18) |
| `update remote-to-local` | #16 (`[Brew] update remote-to-local --module=brew` — first module; dotfiles direction is #17) |
| `status` | #11 |
| `doctor` | #10 |

Note: `update {local,remote}-to-remote` are split per-module on GitHub rather than per-command. The stub cites the brew-module issue since brew is the first module landing; users following the link will see the full update epic from there.

Groups (`brew`, `update`) use `Commands:` to nest their two children; the parent command itself has no `Action`, so `macheim brew` with no subcommand shows that group's help.

## 7. `main.go`

```go
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
```

The previous file's three `_ = X` placeholder lines are gone — those variables now feed `Runtime`.

## 8. `cmd/root_test.go`

Integration tests using `cli.Command.Run` directly. Captures stdout via `cmd.Writer = &bytes.Buffer{}`.

Cases:

| Test | Input | Expected |
|---|---|---|
| `Test_Help_ListsEveryCommand` | `[]string{"macheim", "--help"}` | stdout contains each of: bootstrap, brew, zsh, dotfiles, macos, downloads, update, status, doctor |
| `Test_Version_ShowsBuildIdentity` | `[]string{"macheim", "--version"}` | stdout contains `v0.0.1 (commit deadbeef, built 2026-01-01T00:00:00Z)` (matches the test's fixed rt) |
| `Test_Stub_DoctorPrintsNotImplemented` | `[]string{"macheim", "doctor"}` | stdout = `doctor: not implemented yet (see issue #10)\n`; err = nil |
| `Test_Stub_BrewBundle` | `[]string{"macheim", "brew", "bundle"}` | stdout = `brew bundle: not implemented yet (see issue #13)\n`; err = nil |
| `Test_Stub_UpdateLocalToRemote` | `[]string{"macheim", "update", "local-to-remote"}` | stdout = `update local-to-remote: not implemented yet (see issue #18)\n` |
| `Test_FlagValidation_VerboseAndQuiet` | `[]string{"macheim", "--verbose", "--quiet", "doctor"}` | err != nil; err.Error() contains "mutually exclusive" |
| `Test_FlagInheritance_DryRunSetInRuntime` | `[]string{"macheim", "--dry-run", "doctor"}` | after run, `rt.DryRun == true` |

A helper `runRoot(t, args ...)` constructs a fresh `*Runtime`, builds the root, redirects `Writer`/`ErrWriter` to buffers, and returns `(stdout, stderr, err, rt)`. All seven cases use it.

## 9. Behavior notes (for the implementer)

- **`Before` returns `(context.Context, error)` in urfave/cli/v3** — different from v2. The implementer should rely on the package docs at the time of writing rather than memory.
- **Group commands with no `Action`** — urfave/cli/v3 shows the group's help and exits 2 when invoked with no subcommand. That's fine for `macheim brew` / `macheim update`; no special handling needed.
- **`EnableShellCompletion: true`** auto-registers the `completion` subcommand. No `cmd/completion.go` needed for this PR. Sub-issue #38 (referenced from issue-list scan) ships completion polish later.
- **`MACHEIM_REPO` env var fallback** — the `--repo` flag's `Sources: cli.EnvVars("MACHEIM_REPO")` covers step 2 of GOALS.md's repo discovery chain. Steps 3-5 (config.yaml, conventions, embed fallback) land in sub-issue #6.

## 10. Testing strategy

- Unit: `runtime_test.go` covers `Validate` and `VersionString`.
- Integration: `root_test.go` covers help, version, every stub, flag validation, flag inheritance into Runtime.
- No fixtures, no shell out, no environment dependencies — `go test -race ./...` is the only command.
- Expected: `make lint` clean (no new linter violations); the existing 9 linters cover the surface.

## 11. Risk assessment

| Risk | Likelihood | Mitigation |
|---|---|---|
| `tools.go` becomes redundant — `urfave/cli/v3` is now a real import | High (by design) | Delete the blank `_ "github.com/urfave/cli/v3"` line from `tools.go` once `cmd/root.go` lands; leave the `yaml.v3` blank import (not yet used in real code). Re-run `go mod tidy`. |
| `Before` signature change between minor versions of urfave/cli/v3 | Low | We pinned v3.9.0; the Makefile already uses `-trimpath`; CI (sub-issue #23) will catch regressions on bumps. |
| `notImplemented` issue numbers wrong | Low | Each is verified against the issue list during the implementation commit. If an issue is renumbered (e.g., split), update in a follow-up commit. |
| `EnableShellCompletion` interferes with help output | Low | Covered by `Test_Help_ListsEveryCommand` — failure is immediate and obvious. |
| Concurrent `*Runtime` reads/writes | Very low | CLI is single-goroutine: `Before` runs to completion before any subcommand `Action`. No locks needed. |

## 12. Backward compatibility

N/A — first user-visible surface. Future flag changes will require a deprecation cycle per `Code.md §9.3`.

## 13. Dependencies (on other sub-issues)

- **Blocked by:** #4 (done — merged in PR #102)
- **Blocks:** every later sub-issue that lands an actual subcommand body. The stub file each lives in is the implementation site.

## 14. Out of scope (deferred)

- Actual command behavior — every command body is a stub returning the "not implemented yet" line.
- Config discovery beyond the `--repo` flag and `MACHEIM_REPO` env var — #6.
- Output helper / ANSI suppression infrastructure — folded into the first command that emits colored output (likely #10 or a small precursor).
- Shell completion polish (custom completions per-flag) — #38.

## 15. Acceptance verification

Maps each #5 AC to a verifying command:

| AC | Verifying command | Expected |
|---|---|---|
| `cmd/root.go` returns `*cli.Command` | `grep 'func NewRoot' cmd/root.go` | one match |
| All GOALS.md commands registered | `./dist/macheim --help` | stdout lists bootstrap, brew, zsh, dotfiles, macos, downloads, update, status, doctor, completion |
| Inheritable global flags | `./dist/macheim --help` | stdout shows --repo, --dry-run, --verbose, --quiet, --yes, --no-color |
| Before hook parses into Runtime | `go test ./cmd -run Test_FlagInheritance_DryRunSetInRuntime` | exit 0 |
| Stubs print "not implemented yet" | `./dist/macheim doctor` | stdout = `doctor: not implemented yet (see issue #10)`; exit 0 |
| `--help`, `brew --help`, `update --help` | `./dist/macheim brew --help && ./dist/macheim update --help` | each shows two-child subtree |
| `--no-color` suppresses ANSI | (no ANSI exists yet) | trivially passes; verified in PR following first ANSI emitter |

Plus the regression checks inherited from #4:

| Check | Command | Expected |
|---|---|---|
| Build | `make build` | exit 0; `dist/macheim` exists |
| Tests | `make test` | exit 0; new tests in `cmd/` and `internal/config/` run |
| Lint | `make lint` | `0 issues` |

## 16. Commit plan

One PR, four commits (each independently passes `make lint && make test`):

1. `feat(config): add Runtime struct with Validate + VersionString` — `internal/config/runtime.go` + `runtime_test.go`
2. `feat(cmd): add root command with global flags and Before hook` — `cmd/root.go` (no subcommands yet; --version + Before validation work; root_test.go for those two cases)
3. `feat(cmd): register every GOALS.md command as a stub` — every `cmd/<name>.go` + remaining stub tests in `root_test.go`
4. `feat: wire cmd.NewRoot into main; drop redundant tools.go entry` — `main.go` update, `tools.go` drops the `urfave/cli/v3` blank import (it's a real import now), `go mod tidy` run

Closing line on commit 4:

```
Closes #5
```
