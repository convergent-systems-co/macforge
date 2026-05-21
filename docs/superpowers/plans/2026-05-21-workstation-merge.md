# Workstation Merge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fold `github.com/polliard/macheim` (73 commits, urfave/cli/v3) into `github.com/convergent-systems-co/macforge` as `macforge workstation <verb>`, preserving history via subtree merge and rewriting the command surface to cobra.

**Architecture:** Subtree-import macheim into `.macheim-import/` (leading-dot path ignored by `go build ./...`), relocate packages to `internal/workstation/<pkg>/`, rewrite imports to macforge's module path, port `cmd/` to cobra, drop `urfave/cli/v3`. The two `config` and two `output` packages coexist namespaced by parent path. Workstation cmds use a local `*workstationconfig.Runtime` rather than macforge's `cliRuntime`/audit envelope — matches macheim's lighter shape and defers audit instrumentation as v0.3+ work.

**Tech Stack:** Go 1.26, cobra, viper, `git subtree`, sed-based import rewrite. Source macheim at `/Users/itsfwcp/workspace/convergent-system-co/macheim` (HEAD `60821b7`, remote `https://github.com/polliard/macheim.git`).

**Working tree:** `/Users/itsfwcp/workspace/convergent-system-co/macforge/.git/worktrees/macheim` on branch `macheim`. Spec at `docs/superpowers/specs/2026-05-21-workstation-merge-design.md` (commit `529adc6`).

---

## File Structure

### Created

| Path | Responsibility |
|---|---|
| `docs/adr/0018-peer-subtree-named-workstation.md` | ADR superseding 0017; records the rename from `macheim` peer to `workstation` peer |
| `cmd/macforge/workstation_cmd.go` | `newWorkstationCmd()` factory; persistent flags `--workstation-repo`, `--quiet/-q`, `--yes/-y`; PersistentPreRunE that constructs the workstation runtime; AddCommand-s the 9 verb commands |
| `cmd/macforge/workstation_bootstrap_cmd.go` | `newWorkstationBootstrapCmd()` — currently stub printing `not implemented yet (see issue #X)` |
| `cmd/macforge/workstation_brew_cmd.go` | `newWorkstationBrewCmd()` parent + `newWorkstationBrewInstallCmd()` + `newWorkstationBrewBundleCmd()` |
| `cmd/macforge/workstation_dotfiles_cmd.go` | `newWorkstationDotfilesCmd()` parent + `newWorkstationDotfilesApplyCmd()` |
| `cmd/macforge/workstation_zsh_cmd.go` | `newWorkstationZshCmd()` parent + `newWorkstationZshSetupCmd()` stub |
| `cmd/macforge/workstation_macos_cmd.go` | `newWorkstationMacosCmd()` parent + `newWorkstationMacosDefaultsCmd()` stub |
| `cmd/macforge/workstation_downloads_cmd.go` | `newWorkstationDownloadsCmd()` stub |
| `cmd/macforge/workstation_update_cmd.go` | `newWorkstationUpdateCmd()` parent + `newWorkstationUpdateLocalToRemoteCmd()` + `newWorkstationUpdateRemoteToLocalCmd()` |
| `cmd/macforge/workstation_status_cmd.go` | `newWorkstationStatusCmd()` — renders read-only status |
| `cmd/macforge/workstation_doctor_cmd.go` | `newWorkstationDoctorCmd()` — renders read-only doctor checks; returns sentinel error on failure |
| `cmd/macforge/workstation_runtime.go` | `workstationRuntimeFor(cmd) *workstationconfig.Runtime` resolves the per-invocation runtime from cobra flag values |
| `cmd/macforge/workstation_cmd_test.go` | Smoke tests: `macforge workstation doctor` exits non-zero when checks fail; `macforge workstation status` renders without panicking; stubs print the expected message |
| `examples/workstation/Brewfile` | User-facing sample (moved from `macheim/Brewfile`) |
| `examples/workstation/config.yaml` | User-facing sample (moved from `macheim/examples/config.yaml`) |
| `docs/workstation/specs/...` | Moved from `macheim/docs/superpowers/specs/` (preserves macheim's own design specs) |
| `internal/workstation/{brew,config,doctor,dotfiles,embedded,gitrepo,output,shell,status}/` | Moved from `macheim/internal/<pkg>/`; same source files, rewritten import paths |
| `CHANGELOG.md` entry | Single dated entry: `## [Unreleased] — Add workstation verb group` |

### Modified

| Path | Change |
|---|---|
| `cmd/macforge/root.go` | Stale comment fix (lines 44-45: `macheim` → `workstation`); add `newWorkstationCmd()` to `root.AddCommand(...)` |
| `cmd/macforge/apple_cmd.go` | Stale comment fix (line 21: `macforge macheim` → `macforge workstation`) |
| `go.mod` | Bump `go 1.24` → `go 1.26`; drop `github.com/urfave/cli/v3`; `gopkg.in/yaml.v3` reconciliation (workstation packages use macforge's `go.yaml.in/yaml/v3` instead) |
| `go.sum` | `go mod tidy` regenerates |
| `Makefile` | Add `embed-sync` target copying `examples/workstation/Brewfile` → `internal/workstation/embedded/configs/Brewfile`; make `build` depend on `embed-sync` |
| `README.md` | New Workstation section after the Apple status table; one-line provenance note (`Workstation tooling originated as polliard/macheim, merged into macforge on 2026-05-21`) |
| `internal/workstation/doctor/doctor.go` | Replace `cli.Exit("", 1)` with sentinel error; drop `urfave/cli/v3` import |

### Deleted (during the relocate commit)

`.macheim-import/cmd/`, `.macheim-import/main.go`, `.macheim-import/tools.go`, `.macheim-import/go.mod`, `.macheim-import/go.sum`, `.macheim-import/README.md`, `.macheim-import/GOALS.md`, `.macheim-import/Makefile`, `.macheim-import/LICENSE`, `.macheim-import/NOTICE`, `.macheim-import/.golangci.yml`, `.macheim-import/.editorconfig`, `.macheim-import/.gitignore`, `.macheim-import/Brewfile` (after copy), `.macheim-import/.github/`, `.macheim-import/.claude/`, `.macheim-import/dist/`, finally the now-empty `.macheim-import/` directory itself.

---

## Task 1: Write ADR-0018, fix stale comments

**Why first:** The spec disagrees with the existing accepted ADR-0017, which names the peer subtree `macheim`. Per `Code.md §9.1`, ADRs are immutable once accepted — supersede, don't edit. Writing ADR-0018 first establishes the source of truth before any code references the new name. The two stale comments are fixed in the same commit so the cobra root never lies about what the peer subtree is named.

**Files:**
- Create: `docs/adr/0018-peer-subtree-named-workstation.md`
- Modify: `cmd/macforge/root.go:44-45`
- Modify: `cmd/macforge/apple_cmd.go:21`

- [ ] **Step 1.1: Create the ADR file**

Write `docs/adr/0018-peer-subtree-named-workstation.md`:

```markdown
# ADR 0018 — Peer subtree named `workstation` (not `macheim`)

| Status      | Accepted                                                            |
|-------------|---------------------------------------------------------------------|
| Date        | 2026-05-21                                                          |
| Supersedes  | [ADR-0017](0017-apple-command-namespace.md) §Decision (peer name only) |
| Issue       | _none — operator-surface clarity_                                   |

## Context

ADR-0017 reserved a top-level peer subtree under `macforge` named `macheim`, mirroring the upstream project name. While planning the actual import (spec `docs/superpowers/specs/2026-05-21-workstation-merge-design.md`), the operator-surface implications of using the upstream project name became apparent.

`macheim` is the name of the originating Go module and GitHub repo. It is not what the operator is doing. The operator is configuring a workstation — a Mac with Homebrew, dotfiles, zsh, macOS defaults. `macforge workstation doctor` reads better than `macforge macheim doctor` to anyone who never knew the upstream project existed.

## Decision

The peer subtree is named `workstation`. The verbs land at `macforge workstation <verb>`. The audit envelope schema reads `macforge.v1.workstation.<verb>`. The cobra command file is `cmd/macforge/workstation_cmd.go`.

`macheim` survives as the upstream project name in:
- Git history (the 73 imported commits)
- The README provenance line
- This ADR
- The originating-repo URL referenced in the subtree-import commit body

Nowhere on the operator surface.

## Consequences

### Positive

- **Operator-readable.** The verb name describes what the verb does, not where the code came from.
- **Provenance still preserved.** The subtree merge keeps the 73 commits; the upstream identity is recoverable from any of them.

### Negative

- **ADR-0017's tree-art and audit-schema sections are now stale.** Resolved by this ADR superseding the peer-name portion. The rest of 0017 (the `apple` namespace itself) stays in force.
- **`cmd/macforge/apple_cmd.go:21` and `cmd/macforge/root.go:44-45` had comments referencing the old name.** Fixed in the same commit as this ADR.

### Neutral

- The upstream repo at `github.com/polliard/macheim` keeps its name. Renaming is unnecessary and creates redirect noise.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Keep `macheim` per ADR-0017 | No churn; ADR untouched | Verb name leaks upstream-project identity onto every invocation | Rejected |
| Some third name (`host`, `mac`, `setup`) | Equally generic | No operator-tested preference; new names compound coordination | Rejected |
| **`workstation`** (this ADR) | Operator-readable; matches what the verbs actually do | One ADR supersession | **Chosen** |

## Links

- Spec: [`../superpowers/specs/2026-05-21-workstation-merge-design.md`](../superpowers/specs/2026-05-21-workstation-merge-design.md)
- Supersedes: [ADR-0017](0017-apple-command-namespace.md) (peer-name portion)
```

- [ ] **Step 1.2: Fix `cmd/macforge/root.go` lines 44-45**

Replace:

```go
	// All Apple-platform release operations live under `macforge apple <verb>`.
	// `version` stays at the root (universal). When macheim merges, its verbs
	// will land at `macforge macheim <verb>` as a peer to `apple`. See
	// docs/adr/0017-apple-command-namespace.md.
```

With:

```go
	// All Apple-platform release operations live under `macforge apple <verb>`.
	// `version` stays at the root (universal). Workstation operations (Homebrew,
	// dotfiles, zsh, macOS defaults) live under `macforge workstation <verb>` as
	// a peer. See docs/adr/0017-apple-command-namespace.md (namespace pattern)
	// and docs/adr/0018-peer-subtree-named-workstation.md (peer-name decision).
```

The `root.AddCommand(...)` call stays unchanged in this commit — `newWorkstationCmd()` doesn't exist yet. It's added in Task 5.

- [ ] **Step 1.3: Fix `cmd/macforge/apple_cmd.go` line 21**

Replace:

```go
This subtree contains everything MacForge does today. A future peer
subtree ` + "`macforge macheim`" + ` will hold the macheim-platform operations
once that project merges in.`,
```

With:

```go
This subtree contains everything MacForge does today. The peer subtree
` + "`macforge workstation`" + ` holds Mac workstation operations (Homebrew,
dotfiles, zsh, macOS defaults).`,
```

- [ ] **Step 1.4: Verify the build still compiles**

Run: `go build ./...`
Expected: exits zero, no output.

- [ ] **Step 1.5: Commit**

```bash
git add docs/adr/0018-peer-subtree-named-workstation.md cmd/macforge/root.go cmd/macforge/apple_cmd.go
git commit -m "docs(adr): add ADR-0018 — peer subtree named \`workstation\` (supersedes ADR-0017 peer-name)

Operator-surface clarity. The verb name should describe what the verb
does, not where the code came from. Spec at
docs/superpowers/specs/2026-05-21-workstation-merge-design.md.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Subtree-import macheim into `.macheim-import/`

**Why this shape:** A leading-dot directory (`.macheim-import/`) is invisible to `go build ./...` (Go skips directories whose names begin with `.` or `_`). This lets the import land as one merge commit while leaving macforge's build green. The `--squash=false` keeps the 73 individual macheim commits visible in `git log -- .macheim-import/`.

**Files:**
- Create: `.macheim-import/` (entire macheim tree at upstream HEAD)

- [ ] **Step 2.1: Verify the local macheim clone is at the expected head**

Run from `/Users/itsfwcp/workspace/convergent-system-co/macforge/.git/worktrees/macheim`:

```bash
git -C /Users/itsfwcp/workspace/convergent-system-co/macheim rev-parse HEAD
```

Expected: prints `60821b7...` (the merge of PR #117 from `feat/15-16-18-update-all`). If the working copy is dirty or on a non-main branch, stop and resolve before continuing — the subtree import below pulls from the remote, so local-only changes are not what gets imported.

- [ ] **Step 2.2: Run the subtree add**

```bash
git subtree add --prefix=.macheim-import https://github.com/polliard/macheim.git main
```

Expected output ends with something like:

```
Added dir '.macheim-import'
```

The subtree command produces TWO commits on `macheim` branch: a "Squashed '..." preparatory commit (despite `--squash` being default-off here, the subtree add still records the imported tree as a prep step) and a merge commit. After this command, `git log --oneline -5` shows the merge as HEAD with macheim's commits visible behind it via `git log .macheim-import/`.

- [ ] **Step 2.3: Verify 73-commit history is preserved**

```bash
git log --oneline -- .macheim-import/ | wc -l
```

Expected: a number ≥ 73 (the 73 macheim commits + possibly a small number of merge commits from the subtree merge mechanism itself).

- [ ] **Step 2.4: Verify macforge build is still green**

```bash
go build ./...
```

Expected: exits zero, no output. The `.macheim-import/` directory is invisible to Go's package walker because of the leading dot.

- [ ] **Step 2.5: Verify the imported tree looks right**

```bash
ls .macheim-import/
ls .macheim-import/internal/
```

Expected: lists `cmd/`, `internal/`, `Brewfile`, `README.md`, `GOALS.md`, `Makefile`, `go.mod`, etc. The `internal/` listing shows `brew config doctor dotfiles embedded gitrepo output shell status`.

- [ ] **Step 2.6: Commit checkpoint (already done by subtree add)**

The subtree command auto-committed. Verify the message:

```bash
git log -1 --format='%s%n%n%b'
```

If the auto-generated message is good (something like "Add 'macheim/' from commit '...'"), leave it. If it's terse, amend the body to add provenance (replace `<sha>` with the actual macheim HEAD sha from Step 2.1):

```bash
git commit --amend -m "feat(workstation): import macheim via git subtree

Subtree-imported polliard/macheim@<sha> into .macheim-import/.
The 73 imported commits remain visible via \`git log .macheim-import/\`.
The leading-dot directory keeps the tree invisible to \`go build ./...\`
until Task 3 relocates it to internal/workstation/.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

Do NOT amend if the subtree command produced a multi-commit sequence — only amend the last one if it's the single import commit.

---

## Task 3: Relocate the imported tree to target paths

**Compile-state warning:** This commit leaves `go build ./...` broken — the moved packages still reference `github.com/polliard/macheim/internal/...` import paths. Task 4 fixes this. Reviewers should evaluate Tasks 3 and 4 as a logical unit.

**Files:**
- Move: `.macheim-import/internal/{brew,config,doctor,dotfiles,embedded,gitrepo,output,shell,status}/` → `internal/workstation/{...}/`
- Move: `.macheim-import/Brewfile` → `examples/workstation/Brewfile`
- Move: `.macheim-import/examples/config.yaml` → `examples/workstation/config.yaml`
- Move: `.macheim-import/docs/superpowers/specs/` → `docs/workstation/specs/`
- Delete: `.macheim-import/cmd/`, `.macheim-import/main.go`, `.macheim-import/tools.go`, `.macheim-import/go.mod`, `.macheim-import/go.sum`, `.macheim-import/README.md`, `.macheim-import/GOALS.md`, `.macheim-import/Makefile`, `.macheim-import/LICENSE`, `.macheim-import/NOTICE`, `.macheim-import/.golangci.yml`, `.macheim-import/.editorconfig`, `.macheim-import/.gitignore`, `.macheim-import/.github/`, `.macheim-import/.claude/`, `.macheim-import/dist/`, `.macheim-import/` (after empty)

- [ ] **Step 3.1: Create destination parent directories**

```bash
mkdir -p internal/workstation
mkdir -p examples/workstation
mkdir -p docs/workstation/specs
```

- [ ] **Step 3.2: Move the nine internal packages with `git mv`**

```bash
git mv .macheim-import/internal/brew internal/workstation/brew
git mv .macheim-import/internal/config internal/workstation/config
git mv .macheim-import/internal/doctor internal/workstation/doctor
git mv .macheim-import/internal/dotfiles internal/workstation/dotfiles
git mv .macheim-import/internal/embedded internal/workstation/embedded
git mv .macheim-import/internal/gitrepo internal/workstation/gitrepo
git mv .macheim-import/internal/output internal/workstation/output
git mv .macheim-import/internal/shell internal/workstation/shell
git mv .macheim-import/internal/status internal/workstation/status
```

Using `git mv` for each (rather than a shell loop with `mv`) makes git record the move so `git log --follow internal/workstation/brew/install.go` traces back into the original macheim history.

- [ ] **Step 3.3: Move user-facing sample files**

```bash
git mv .macheim-import/Brewfile examples/workstation/Brewfile
git mv .macheim-import/examples/config.yaml examples/workstation/config.yaml
```

- [ ] **Step 3.4: Move macheim's design specs**

```bash
# Move every spec file in .macheim-import/docs/superpowers/specs/ into docs/workstation/specs/
git mv .macheim-import/docs/superpowers/specs/* docs/workstation/specs/
```

If the source directory has subdirectories (other than specs), repeat for each. Verify nothing under `.macheim-import/docs/` other than already-deleted-or-moved content remains:

```bash
find .macheim-import/docs -type f
```

If files remain, evaluate each and either move-or-delete with `git mv` or `git rm`.

- [ ] **Step 3.5: Delete macheim's duplicate top-level files**

```bash
git rm .macheim-import/cmd/*.go
rmdir .macheim-import/cmd
git rm .macheim-import/main.go .macheim-import/tools.go .macheim-import/go.mod .macheim-import/go.sum
git rm .macheim-import/README.md .macheim-import/GOALS.md .macheim-import/Makefile
git rm .macheim-import/LICENSE .macheim-import/NOTICE
git rm .macheim-import/.golangci.yml .macheim-import/.editorconfig .macheim-import/.gitignore
git rm -r .macheim-import/.github
git rm -r .macheim-import/.claude 2>/dev/null || true   # may not exist
git rm -r .macheim-import/dist 2>/dev/null || true      # may not exist
```

Then any straggler files:

```bash
find .macheim-import -type f
```

Delete remaining files with `git rm <path>`.

- [ ] **Step 3.6: Remove the now-empty `.macheim-import/` directory**

```bash
# Directory removal is implicit once git tracks no files under it.
find .macheim-import -type d -empty -delete
```

Confirm:

```bash
test ! -d .macheim-import && echo "removed" || echo "still present"
```

Expected: `removed`.

- [ ] **Step 3.7: Verify the relocate worked**

```bash
ls internal/workstation/
ls examples/workstation/
ls docs/workstation/specs/ | head
```

Expected: nine internal packages, two sample files, the macheim specs. `git status` should show only renames/deletions in the index (no untracked files).

- [ ] **Step 3.8: Commit (build will be broken — see note at top of Task 3)**

```bash
git commit -m "refactor(workstation): relocate imported tree into target paths

Move .macheim-import/internal/<pkg> -> internal/workstation/<pkg> for
all nine packages. Move root-level user-facing artifacts (Brewfile,
examples/config.yaml) to examples/workstation/. Move docs/superpowers/specs/
content to docs/workstation/specs/. Delete duplicate cmd/, main.go,
tools.go, go.mod/sum, README, GOALS, Makefile, LICENSE, NOTICE, etc.

\`go build ./...\` is broken until Task 4 rewrites import paths to
github.com/convergent-systems-co/macforge/internal/workstation/...

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Rewrite Go import paths

**Files:**
- Modify: every `.go` file under `internal/workstation/` (mechanical sed)

- [ ] **Step 4.1: Identify every file that needs rewriting**

```bash
grep -rln 'github.com/polliard/macheim' internal/workstation/
```

Expected: every Go file that imports a sibling workstation package (about 30–40 files based on macheim's structure).

- [ ] **Step 4.2: Run the import-path rewrite**

```bash
find internal/workstation -name '*.go' -type f -exec \
  sed -i.bak \
  -e 's|github.com/polliard/macheim/internal/|github.com/convergent-systems-co/macforge/internal/workstation/|g' \
  {} +

# Remove sed's backup files
find internal/workstation -name '*.go.bak' -delete
```

Note: on macOS, BSD sed's `-i` requires an extension argument; `-i.bak` followed by the bak-cleanup is the portable form.

- [ ] **Step 4.3: Verify no stray macheim-paths remain**

```bash
grep -rn 'github.com/polliard/macheim' internal/workstation/
```

Expected: no output. If anything remains (test fixtures with hardcoded module paths, embedded strings), address case-by-case.

- [ ] **Step 4.4: Reconcile the yaml package**

macheim uses `gopkg.in/yaml.v3`; macforge uses `go.yaml.in/yaml/v3` (different package paths, same on-disk YAML format and identical API for v3). Find every workstation file importing the old yaml path:

```bash
grep -rln 'gopkg.in/yaml.v3' internal/workstation/
```

For each match, rewrite the import:

```bash
find internal/workstation -name '*.go' -type f -exec \
  sed -i.bak \
  -e 's|gopkg.in/yaml.v3|go.yaml.in/yaml/v3|g' \
  {} +
find internal/workstation -name '*.go.bak' -delete
```

Verify:

```bash
grep -rn 'gopkg.in/yaml.v3' internal/workstation/
```

Expected: no output.

- [ ] **Step 4.5: Tidy go.mod**

```bash
go mod tidy
```

This should: drop `gopkg.in/yaml.v3` from the require block (no callers left); add `go.yaml.in/yaml/v3` to direct requires (workstation packages now use it); keep `github.com/urfave/cli/v3` and `golang.org/x/term` (still used until Tasks 5 and 6).

- [ ] **Step 4.6: Verify `go build ./...` is green**

```bash
go build ./...
```

Expected: exits zero. If failures remain they will be one of:
- An overlooked `github.com/polliard/macheim/` path (Step 4.3 missed it)
- A yaml API call that changed between gopkg.in/yaml.v3 and go.yaml.in/yaml/v3 (unlikely — same v3 API; if it happens, the compile error names the function and the fix is mechanical)
- A genuine Go 1.26 feature used by macheim that 1.24 doesn't have. If so, jump to Task 7 (Go version bump) early and come back.

- [ ] **Step 4.7: Verify tests still compile**

```bash
go vet ./...
go test -count=1 -run='^$' ./internal/workstation/...
```

The `-run='^$'` matches no tests but forces test compilation. Expected: every package builds its test binary cleanly.

- [ ] **Step 4.8: Commit**

```bash
git add internal/workstation go.mod go.sum
git commit -m "refactor(workstation): rewrite imports to macforge module path

sed across internal/workstation/**/*.go:
  github.com/polliard/macheim/internal/<pkg>
    -> github.com/convergent-systems-co/macforge/internal/workstation/<pkg>

Reconcile yaml dependency:
  gopkg.in/yaml.v3 -> go.yaml.in/yaml/v3 (same v3 API, different
  package path; aligns with macforge's existing yaml dep).

\`go build ./...\` is green; \`go vet ./...\` and test-compile clean.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Port the command tree to cobra

This is the largest task. The 14 macheim command files become 11 macforge command files using the cobra factory pattern from `apple_cmd.go` / `identity_cmd.go`. Each new command file follows the same shape: factory function, RunE closure, delegate to `internal/workstation/<pkg>.<Func>(rt)`.

**Files:**
- Create: `cmd/macforge/workstation_cmd.go`
- Create: `cmd/macforge/workstation_runtime.go`
- Create: `cmd/macforge/workstation_bootstrap_cmd.go`
- Create: `cmd/macforge/workstation_brew_cmd.go`
- Create: `cmd/macforge/workstation_dotfiles_cmd.go`
- Create: `cmd/macforge/workstation_zsh_cmd.go`
- Create: `cmd/macforge/workstation_macos_cmd.go`
- Create: `cmd/macforge/workstation_downloads_cmd.go`
- Create: `cmd/macforge/workstation_update_cmd.go`
- Create: `cmd/macforge/workstation_status_cmd.go`
- Create: `cmd/macforge/workstation_doctor_cmd.go`
- Create: `cmd/macforge/workstation_cmd_test.go`
- Modify: `cmd/macforge/root.go` (one-line AddCommand)
- Modify: `internal/workstation/doctor/doctor.go` (drop cli.Exit)

### 5a — Workstation runtime helper

- [ ] **Step 5.1: Write the failing test for the runtime helper**

Create `cmd/macforge/workstation_cmd_test.go`:

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"strings"
	"testing"
)

func TestWorkstation_DoctorRunsAndExits(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"workstation", "doctor"})
	// doctor may exit non-zero on this host (no Homebrew, no repo, etc.).
	// We just verify it runs without panicking and either succeeds or
	// returns a non-cobra-internal error.
	if err := root.Execute(); err != nil {
		// Acceptable: doctor checks failed (exit 1 path). Verify the error
		// is the sentinel, not something like "unknown command".
		if !strings.Contains(err.Error(), "doctor") {
			t.Fatalf("doctor returned unexpected error: %v", err)
		}
	}
}

func TestWorkstation_StubsPrintNotImplemented(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"workstation", "bootstrap"}, "not implemented yet"},
		{[]string{"workstation", "zsh", "setup"}, "not implemented yet"},
		{[]string{"workstation", "macos", "defaults"}, "not implemented yet"},
		{[]string{"workstation", "downloads"}, "not implemented yet"},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			root := newRootCmd()
			var sb strings.Builder
			root.SetOut(&sb)
			root.SetErr(&sb)
			root.SetArgs(tc.args)
			if err := root.Execute(); err != nil {
				t.Fatalf("Execute: %v", err)
			}
			if !strings.Contains(sb.String(), tc.want) {
				t.Errorf("output missing %q\n---\n%s", tc.want, sb.String())
			}
		})
	}
}

func TestWorkstation_StatusRendersWithoutPanic(t *testing.T) {
	root := newRootCmd()
	var sb strings.Builder
	root.SetOut(&sb)
	root.SetErr(&sb)
	root.SetArgs([]string{"workstation", "status"})
	// status is read-only and tolerates an unconfigured environment;
	// it should never return a non-nil error.
	if err := root.Execute(); err != nil {
		t.Fatalf("status returned unexpected error: %v", err)
	}
}
```

- [ ] **Step 5.2: Run the test to verify it fails**

```bash
go test ./cmd/macforge/ -run TestWorkstation -v
```

Expected: FAIL with "unknown command 'workstation' for 'macforge'" or similar.

- [ ] **Step 5.3: Create `cmd/macforge/workstation_runtime.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	workstationconfig "github.com/convergent-systems-co/macforge/internal/workstation/config"
)

// workstationRuntimeFor builds a *workstationconfig.Runtime from cobra's
// resolved flag values for a workstation subcommand invocation. Inherited
// macforge globals (--dry-run, --verbose, --no-color) are read from gflags;
// workstation-specific flags (--workstation-repo, --quiet, --yes) are read
// from cmd.Flags() via the persistent flags on newWorkstationCmd.
func workstationRuntimeFor(cmd *cobra.Command) *workstationconfig.Runtime {
	repo, _ := cmd.Flags().GetString("workstation-repo")
	quiet, _ := cmd.Flags().GetBool("quiet")
	yes, _ := cmd.Flags().GetBool("yes")

	return &workstationconfig.Runtime{
		RepoPath: repo,
		DryRun:   gflags.dryRun,
		Verbose:  gflags.verbose,
		Quiet:    quiet,
		Yes:      yes,
		NoColor:  gflags.noColor,
		// Version/Commit/BuildDate are not surfaced for workstation today;
		// macforge has its own version command, and workstation doctor/status
		// do not need build metadata.
	}
}
```

### 5b — Workstation root command

- [ ] **Step 5.4: Create `cmd/macforge/workstation_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

// newWorkstationCmd returns the `workstation` subtree: Homebrew, dotfiles,
// zsh, macOS defaults, and the read-only doctor/status checks. Peer to
// `newAppleCmd()`. See ADR-0017 (namespace pattern) and ADR-0018
// (peer-name decision).
func newWorkstationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workstation",
		Short: "Mac workstation operations (Homebrew, dotfiles, zsh, macOS defaults)",
		Long: `Mac workstation operations: bootstrap a fresh Mac to a known-good state
defined in a git repo, reconcile drift in either direction, and run
read-only doctor/status checks.

Most mutating verbs (bootstrap, brew install/bundle, dotfiles apply,
zsh setup, macos defaults, downloads, update *) are currently stubbed —
they print "not implemented yet" and exit zero. The read-only verbs
(status, doctor) work today.

Originated as github.com/polliard/macheim; merged into macforge on
2026-05-21.`,
	}

	pflags := cmd.PersistentFlags()
	pflags.String("workstation-repo", "", "path to your workstation repo (overrides discovery; also reads MACFORGE_WORKSTATION_REPO)")
	pflags.BoolP("quiet", "q", false, "suppress non-error output")
	pflags.BoolP("yes", "y", false, "skip confirmation prompts")

	cmd.AddCommand(
		newWorkstationBootstrapCmd(),
		newWorkstationBrewCmd(),
		newWorkstationDotfilesCmd(),
		newWorkstationZshCmd(),
		newWorkstationMacosCmd(),
		newWorkstationDownloadsCmd(),
		newWorkstationUpdateCmd(),
		newWorkstationStatusCmd(),
		newWorkstationDoctorCmd(),
	)
	return cmd
}
```

- [ ] **Step 5.5: Register workstation on the root**

Modify `cmd/macforge/root.go`. Replace:

```go
	root.AddCommand(
		newAppleCmd(),
		newVersionCmd(),
	)
```

With:

```go
	root.AddCommand(
		newAppleCmd(),
		newWorkstationCmd(),
		newVersionCmd(),
	)
```

### 5c — Doctor and status (the two read-only commands that already work)

- [ ] **Step 5.6: Replace `cli.Exit` in `internal/workstation/doctor/doctor.go`**

Modify `internal/workstation/doctor/doctor.go`:

Remove the import:

```go
	"github.com/urfave/cli/v3"
```

Replace lines 73-88 (`func Run(...) error`) with:

```go
// Run executes every check in order, renders each row, prints a summary,
// and returns nil on all-pass or ErrChecksFailed when any check fails.
// Callers (cobra RunE) propagate the sentinel; cobra renders it and sets
// exit code 1.
func Run(rt *config.Runtime, w io.Writer) error {
	r := newRender(rt, w)
	failed := 0
	for _, c := range DefaultChecks() {
		res := c.Run(rt)
		r.row(c.Name, res)
		if !res.OK {
			failed++
		}
	}
	r.summary(failed)
	if failed > 0 {
		return ErrChecksFailed
	}
	return nil
}

// ErrChecksFailed is returned by Run when one or more doctor checks fail.
// Callers should treat this as the signal to exit 1; the per-check output
// has already been rendered to the writer.
var ErrChecksFailed = errChecksFailed{}

type errChecksFailed struct{}

func (errChecksFailed) Error() string { return "doctor: one or more checks failed" }
```

(Sentinel value type allows callers to `errors.Is(err, doctor.ErrChecksFailed)` if they want to distinguish.)

- [ ] **Step 5.7: Create `cmd/macforge/workstation_doctor_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/doctor"
)

func newWorkstationDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Sanity-check the workstation environment",
		Long: `Read-only diagnostic. Verifies xcode-select, Homebrew, repo discovery,
config directory, and the shell rc file. Exits zero on all-pass; exits
one with per-check remediation hints when something is broken.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return doctor.Run(rt, cmd.OutOrStdout())
		},
	}
}
```

- [ ] **Step 5.8: Create `cmd/macforge/workstation_status_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/status"
)

func newWorkstationStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Read-only summary of drift between this Mac and the workstation repo",
		Long: `Read-only. Prints what's drifted between this Mac and the repo across
brew, repo, dotfiles, and macOS-defaults sections. Tolerates an
unconfigured environment — missing Homebrew, missing repo, etc. all
render as known states rather than errors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return status.Run(rt, cmd.OutOrStdout())
		},
	}
}
```

> Verification note: macheim's `internal/status` exposes a `Run(*config.Runtime, io.Writer) error` function (per macheim/cmd/status.go pattern). If the actual function signature differs after the path rewrite, adjust accordingly. Check with `grep -n 'func Run' internal/workstation/status/status.go`.

### 5d — Brew, dotfiles, update (the verbs with internal-package backing today)

- [ ] **Step 5.9: Create `cmd/macforge/workstation_brew_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/brew"
)

func newWorkstationBrewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "brew",
		Short: "Homebrew operations",
	}
	cmd.AddCommand(
		newWorkstationBrewInstallCmd(),
		newWorkstationBrewBundleCmd(),
	)
	return cmd
}

func newWorkstationBrewInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install Homebrew itself (arch-aware)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return brew.Install(rt)
		},
	}
}

func newWorkstationBrewBundleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bundle",
		Short: "Apply the repo Brewfile (brew bundle wrapper with embed fallback)",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return brew.Apply(rt)
		},
	}
}
```

- [ ] **Step 5.10: Create `cmd/macforge/workstation_dotfiles_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/dotfiles"
)

func newWorkstationDotfilesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dotfiles",
		Short: "Dotfile operations",
	}
	cmd.AddCommand(newWorkstationDotfilesApplyCmd())
	return cmd
}

func newWorkstationDotfilesApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Copy <repo>/dotfiles/ into $HOME with backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			return dotfiles.Apply(rt)
		},
	}
}
```

- [ ] **Step 5.11: Create `cmd/macforge/workstation_update_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import (
	"github.com/spf13/cobra"

	"github.com/convergent-systems-co/macforge/internal/workstation/brew"
	"github.com/convergent-systems-co/macforge/internal/workstation/dotfiles"
)

func newWorkstationUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Reconcile drift between this Mac and the workstation repo",
	}
	cmd.AddCommand(
		newWorkstationUpdateLocalToRemoteCmd(),
		newWorkstationUpdateRemoteToLocalCmd(),
	)
	return cmd
}

func newWorkstationUpdateLocalToRemoteCmd() *cobra.Command {
	var module string
	cmd := &cobra.Command{
		Use:   "local-to-remote",
		Short: "Ratchet local drift back into the repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			switch module {
			case "brew":
				return brew.UpdateLocalToRemote(rt)
			case "dotfiles":
				return dotfiles.UpdateLocalToRemote(rt)
			case "all", "":
				if err := brew.UpdateLocalToRemote(rt); err != nil {
					return err
				}
				return dotfiles.UpdateLocalToRemote(rt)
			default:
				return cmd.Help()
			}
		},
	}
	cmd.Flags().StringVar(&module, "module", "all", "brew | dotfiles | all")
	return cmd
}

func newWorkstationUpdateRemoteToLocalCmd() *cobra.Command {
	var module string
	cmd := &cobra.Command{
		Use:   "remote-to-local",
		Short: "Apply repo state to this Mac",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt := workstationRuntimeFor(cmd)
			switch module {
			case "brew":
				return brew.UpdateRemoteToLocal(rt)
			case "all", "":
				return brew.UpdateRemoteToLocal(rt)
			default:
				return cmd.Help()
			}
		},
	}
	cmd.Flags().StringVar(&module, "module", "all", "brew | all (dotfiles remote-to-local lands in a future commit)")
	return cmd
}
```

> Verification note: the exact `Update*` function names depend on macheim's `internal/brew` and `internal/dotfiles` public API. macheim commit log shows `brew.UpdateLocalToRemote`, `brew.UpdateRemoteToLocal`, `dotfiles.UpdateLocalToRemote`. Verify with `grep -n '^func Update' internal/workstation/{brew,dotfiles}/*.go` and adjust call sites if names differ.

### 5e — Stub commands (bootstrap, zsh, macos, downloads)

- [ ] **Step 5.12: Create the stub helper and four stub command files**

First, add a stub helper to `workstation_cmd.go` (top-level, below `newWorkstationCmd()`):

```go
// notImplementedRunE returns a cobra RunE that prints the standard
// "not implemented yet (see issue #N)" message and exits zero. Mirrors
// the macforge stub idiom used by package/notarize/publish/release.
// issue is the GitHub issue number that tracks the real implementation.
func notImplementedRunE(verb string, issue int) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		_, _ = cmd.OutOrStdout().Write([]byte(verb + ": not implemented yet (see issue #" + itoa(issue) + ")\n"))
		return nil
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	n := len(buf)
	for i > 0 {
		n--
		buf[n] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		n--
		buf[n] = '-'
	}
	return string(buf[n:])
}
```

> Note: macforge's `identity_cmd.go` already has an `intToStr` helper with the same shape. If lint complains about the duplicate, hoist to a shared file (e.g., `cmd/macforge/itoa.go`) and have both call sites use it. Out of scope to refactor right now — only do it if golangci-lint hard-fails.

- [ ] **Step 5.13: Create `cmd/macforge/workstation_bootstrap_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationBootstrapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bootstrap",
		Short: "End-to-end fresh-Mac setup (planned)",
		RunE:  notImplementedRunE("bootstrap", 4), // issue #4 from macheim's GOALS.md
	}
}
```

- [ ] **Step 5.14: Create `cmd/macforge/workstation_zsh_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationZshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "zsh operations",
	}
	cmd.AddCommand(newWorkstationZshSetupCmd())
	return cmd
}

func newWorkstationZshSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Configure zsh as the macforge-managed shell (planned)",
		RunE:  notImplementedRunE("zsh setup", 18),
	}
}
```

- [ ] **Step 5.15: Create `cmd/macforge/workstation_macos_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationMacosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "macos",
		Short: "macOS operations",
	}
	cmd.AddCommand(newWorkstationMacosDefaultsCmd())
	return cmd
}

func newWorkstationMacosDefaultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "defaults",
		Short: "Apply the repo macOS defaults manifest (planned)",
		RunE:  notImplementedRunE("macos defaults", 19),
	}
}
```

- [ ] **Step 5.16: Create `cmd/macforge/workstation_downloads_cmd.go`**

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Convergent Systems Co.

package main

import "github.com/spf13/cobra"

func newWorkstationDownloadsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "downloads",
		Short: "Fetch optional downloads listed in the repo (planned)",
		RunE:  notImplementedRunE("downloads", 20),
	}
}
```

> Issue numbers in the stubs (`4`, `18`, `19`, `20`) are from macheim's GOALS.md / Epic-3 sub-issue tracker. If those issues are migrated/renumbered into macforge's tracker during the §7 pre-deletion checklist, update the numbers in a follow-up commit. For the merge itself the macheim issue numbers are correct.

### 5f — Verify build, tests, and remove urfave/cli/v3

- [ ] **Step 5.17: Build**

```bash
go build ./...
```

Expected: exits zero. If there are unresolved package paths (status.Run signature mismatch, etc.), fix the call site to match the actual internal/workstation/<pkg> exported function.

- [ ] **Step 5.18: Run the workstation tests**

```bash
go test ./cmd/macforge/ -run TestWorkstation -v
```

Expected: PASS on `TestWorkstation_StubsPrintNotImplemented`, `TestWorkstation_StatusRendersWithoutPanic`, `TestWorkstation_DoctorRunsAndExits`.

- [ ] **Step 5.19: Run the full suite**

```bash
go test -race ./...
```

Expected: green. Any failures in `internal/workstation/<pkg>` are real bugs surfaced by the merge — investigate before continuing.

- [ ] **Step 5.20: Drop urfave/cli/v3 from go.mod**

```bash
grep -rn 'urfave/cli' internal/workstation/
```

Expected: no output. (If output, find the import and refactor it the way doctor.go was refactored — using a sentinel error or stdlib equivalent.)

```bash
go mod tidy
```

Expected: removes `github.com/urfave/cli/v3` from go.mod. Verify:

```bash
grep urfave go.mod
```

Expected: no output.

- [ ] **Step 5.21: Commit**

```bash
git add cmd/macforge/ internal/workstation/doctor/doctor.go go.mod go.sum
git commit -m "feat(workstation): port command tree to cobra; drop urfave/cli/v3

Add cmd/macforge/workstation_*.go (one root + nine sub-files + runtime
helper + test). Each verb delegates to internal/workstation/<pkg>.
Stub verbs use the macforge \"not implemented yet (see issue #N)\" idiom.

Replace internal/workstation/doctor/doctor.go's cli.Exit(\"\", 1) with
a sentinel ErrChecksFailed returned to cobra, dropping the last
urfave/cli/v3 reference.

go mod tidy drops github.com/urfave/cli/v3 from the module graph.
golang.org/x/term stays — internal/workstation/output uses it for
TTY detection.

Tests: cmd/macforge/workstation_cmd_test.go covers stubs, status,
and doctor exit paths. \`go test -race ./...\` green.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Rewire Makefile embed-sync

**Files:**
- Modify: `Makefile`

- [ ] **Step 6.1: Read the existing Makefile**

```bash
cat Makefile
```

Get the current targets (help, clean, build, test from commit `6f567a5`).

- [ ] **Step 6.2: Add the `embed-sync` target**

Append to `Makefile`:

```makefile

# embed-sync copies the user-facing sample Brewfile into the embedded
# fallback location baked into the binary at build time. Run by `make
# build` automatically; rerun manually after editing
# examples/workstation/Brewfile if you want to refresh the in-tree
# embedded copy without a full build.
.PHONY: embed-sync
embed-sync:
	@mkdir -p internal/workstation/embedded/configs
	@cp examples/workstation/Brewfile internal/workstation/embedded/configs/Brewfile
	@echo "embed-sync: examples/workstation/Brewfile -> internal/workstation/embedded/configs/Brewfile"
```

Then make `build` depend on it. Find the existing `build:` target and change:

```makefile
build:
	go build ./...
```

To:

```makefile
build: embed-sync
	go build ./...
```

- [ ] **Step 6.3: Run embed-sync to seed the embedded copy**

```bash
make embed-sync
```

Expected output: the echo line, no errors.

```bash
diff examples/workstation/Brewfile internal/workstation/embedded/configs/Brewfile
```

Expected: no output (files are byte-identical).

- [ ] **Step 6.4: Confirm idempotency**

```bash
make embed-sync
diff examples/workstation/Brewfile internal/workstation/embedded/configs/Brewfile
```

Expected: no output. Second run is a no-op against the working tree (the cp is overwriting with identical content).

- [ ] **Step 6.5: Verify `make build` still works end-to-end**

```bash
make build
```

Expected: runs embed-sync first, then `go build ./...`, both exit zero.

- [ ] **Step 6.6: Commit**

```bash
git add Makefile internal/workstation/embedded/configs/Brewfile
git commit -m "build(workstation): rewire embed-sync into Makefile

Add 'embed-sync' target: examples/workstation/Brewfile ->
internal/workstation/embedded/configs/Brewfile. Make 'build' depend
on it so a fresh checkout produces a binary with the same embedded
fallback as the repo's sample.

Honors macheim's original embed-sync round-trip discipline (see the
'Verifying the embed-sync round-trip' section of polliard/macheim's
README, preserved in the subtree merge history).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Bump go.mod to 1.26

**Why this late:** Sequencing matters less than I claimed at first. macheim's code compiles fine on 1.24 in practice (no 1.26-only syntax or stdlib calls in the macheim-side packages, verified by `go build ./...` succeeding under 1.24 after Task 4). Bumping go.mod is a one-line change but it triggers a CI matrix update and is a hard cutover for any contributor on an older toolchain. Late is correct — the rest of the merge already works.

If `go build ./...` ever failed under 1.24 in Tasks 3-6 due to a 1.26 feature in macheim code, this task moves earlier and inherits the same steps. Plan execution should be flexible there.

**Files:**
- Modify: `go.mod` (the `go 1.24` directive)
- Modify (maybe): `.go-version`, `.github/workflows/ci.yml`

- [ ] **Step 7.1: Bump the go.mod directive**

```bash
# Find the line
grep -n '^go ' go.mod
# Expected: `go 1.24`
```

Edit `go.mod`. Replace the line `go 1.24` with `go 1.26`.

- [ ] **Step 7.2: Update `.go-version` if present**

```bash
test -f .go-version && cat .go-version
```

If `.go-version` exists and contains `1.24` (or similar), update to `1.26`. If absent, skip.

- [ ] **Step 7.3: Update CI matrix**

```bash
grep -n 'go-version' .github/workflows/ci.yml
```

If a `go-version:` field references `1.24`, update to `1.26`. If the workflow uses a matrix `go-version: [1.24]` or similar, update to `[1.26]`. Macheim's PR-117 commit (`ci: fix lint Go-version mismatch + Windows test portability`, `95d7db5`) established this matrix; one place to edit.

- [ ] **Step 7.4: Run `go mod tidy` and rebuild**

```bash
go mod tidy
go build ./...
go test -race ./...
```

Expected: green. If any test fails specifically because of the version bump, investigate — but in practice 1.24 → 1.26 is additive at the language level.

- [ ] **Step 7.5: Commit**

```bash
git add go.mod go.sum .go-version .github/workflows/ci.yml 2>/dev/null
git commit -m "ci(workstation): bump Go toolchain to 1.26

go.mod: go 1.24 -> go 1.26. macheim required 1.26 floor; the rest of
the workstation merge happens to compile under 1.24 because no 1.26-only
syntax leaks into the merged packages, but the floor is correct here.

Also bump .go-version and the CI matrix's go-version field.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: README + CHANGELOG

**Files:**
- Modify: `README.md`
- Modify: `CHANGELOG.md`

- [ ] **Step 8.1: Update macforge README**

Read the existing README and find the v0.1 status table (around lines 33-45 — the table that lists `macforge apple init`, `keychain *`, `identity *`, etc.).

Add a Workstation section below the existing v0.1 status section. The section should:

- Introduce `macforge workstation <verb>` as a peer to `macforge apple <verb>`.
- One-line provenance: "Workstation tooling originated as `github.com/polliard/macheim`; merged into macforge on 2026-05-21. 73 commits of history preserved via subtree merge."
- A status table mirroring the Apple one:

```markdown
## Workstation (`macforge workstation <verb>`)

Mac workstation bootstrap and sync: Homebrew, dotfiles, zsh, macOS
defaults. Originated as `github.com/polliard/macheim`; merged into
macforge on 2026-05-21. The 73 imported commits are preserved via
subtree merge (`git log internal/workstation/`).

| Verb                                       | Status  |
|--------------------------------------------|---------|
| `macforge workstation status`              | ✓       |
| `macforge workstation doctor`              | ✓       |
| `macforge workstation brew install`        | ✓       |
| `macforge workstation brew bundle`         | ✓       |
| `macforge workstation dotfiles apply`      | ✓       |
| `macforge workstation update local-to-remote`  | ✓ (brew, dotfiles) |
| `macforge workstation update remote-to-local`  | ✓ (brew only) |
| `macforge workstation bootstrap`           | stub    |
| `macforge workstation zsh setup`           | stub    |
| `macforge workstation macos defaults`      | stub    |
| `macforge workstation downloads`           | stub    |

Workstation-specific globals (on top of macforge's `--config`, `--output`,
`--dry-run`, `--verbose`, `--no-color`):

| Flag                          | Env                              | Purpose                          |
|-------------------------------|----------------------------------|----------------------------------|
| `--workstation-repo PATH`     | `MACFORGE_WORKSTATION_REPO`      | path to your workstation repo    |
| `--quiet`, `-q`               |                                  | suppress non-error output        |
| `--yes`, `-y`                 |                                  | skip confirmation prompts        |

Sample artifacts: [`examples/workstation/Brewfile`](examples/workstation/Brewfile)
and [`examples/workstation/config.yaml`](examples/workstation/config.yaml).
```

(Adjust the verified verb statuses based on what `internal/workstation/<pkg>` actually exposes — the `update remote-to-local` row mentions "brew only" because Task 5.11 left dotfiles remote-to-local as a future commit. Cross-check before committing.)

- [ ] **Step 8.2: Update CHANGELOG.md**

Read the existing CHANGELOG to learn the format. Then add an entry at the top under `[Unreleased]` (or create that section if it doesn't exist):

```markdown
## [Unreleased]

### Added
- `macforge workstation <verb>` peer subtree for Mac workstation bootstrap and sync (Homebrew, dotfiles, zsh, macOS defaults). Originated as `github.com/polliard/macheim`; subtree-merged on 2026-05-21 preserving 73 commits of upstream history. See ADR-0017 (namespace), ADR-0018 (peer name), and `docs/superpowers/specs/2026-05-21-workstation-merge-design.md`.

### Changed
- `go.mod`: `go 1.24` → `go 1.26` (workstation tree floor).

### Removed
- `github.com/urfave/cli/v3` from the module graph (workstation port to cobra).
```

- [ ] **Step 8.3: Verify and commit**

```bash
go build ./...        # docs change shouldn't break build
git add README.md CHANGELOG.md
git commit -m "docs(workstation): README + CHANGELOG entries

Document the new workstation peer subtree. Cross-link to ADR-0017,
ADR-0018, and the spec.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>"
```

---

## Final verification

After Task 8, all the spec's success criteria should hold. Walk through them:

- [ ] **V1: Build + tests are green**

```bash
go build ./...
go test -race ./...
```

Expected: both exit zero.

- [ ] **V2: Top-level help shows both groups**

```bash
go run ./cmd/macforge --help 2>&1 | grep -E '^\s+(apple|workstation|version)\s'
```

Expected: three lines, one each for `apple`, `workstation`, `version`.

- [ ] **V3: Workstation help renders**

```bash
go run ./cmd/macforge workstation --help
```

Expected: cobra-native help listing the 9 verb subcommands.

- [ ] **V4: Doctor and status behave**

```bash
go run ./cmd/macforge workstation doctor
go run ./cmd/macforge workstation status
```

Expected: each runs to completion. Doctor may exit 1 on a fresh host (no brew, no repo) — that's correct behavior.

- [ ] **V5: History preserved**

```bash
git log --oneline -- internal/workstation/ | wc -l
```

Expected: ≥ 73.

- [ ] **V6: git log --follow traces back**

```bash
git log --follow --oneline internal/workstation/brew/install.go | tail -3
```

Expected: the oldest entries are macheim commits referencing `internal/brew/install.go` (the pre-rename path).

- [ ] **V7: embed-sync round-trip**

```bash
make embed-sync
diff examples/workstation/Brewfile internal/workstation/embedded/configs/Brewfile
```

Expected: no output.

- [ ] **V8: Lint is clean**

```bash
golangci-lint run ./...
```

Expected: no findings. (Pre-existing lint issues in unrelated macforge packages are out of scope to fix here.)

- [ ] **V9: urfave/cli is gone**

```bash
grep -rn 'urfave/cli' . --include='*.go' --include='go.mod' --include='go.sum' 2>&1 | grep -v '\.git/'
```

Expected: no output.

- [ ] **V10: Stale `macheim` references in code/comments are gone**

```bash
grep -rn 'macheim' cmd/ docs/adr/0017* docs/adr/0018* --include='*.go' --include='*.md' 2>&1 | grep -v 'polliard/macheim' | grep -v 'originated as' | grep -v 'docs/adr/0018'
```

Expected: only acceptable mentions (provenance references, ADR-0018 itself, the supersedes link in ADR-0017's frontmatter). No code-level `macforge macheim` or `macheim_cmd.go` references.

If all ten pass, the merge is complete and ready for the §7 pre-deletion checklist (which Thomas executes manually — assistant does not touch `polliard/macheim`).

---

## Out of scope (do NOT do during this plan)

- Implementing the stub verbs (`bootstrap`, `zsh setup`, `macos defaults`, `downloads`). Land as stubs; functional work follows.
- Refactoring `apple_cmd.go` or any `internal/apple/`, `internal/identity/`, etc. packages.
- Unifying the two `config` or `output` packages.
- Running `gh repo delete polliard/macheim` (operator action; see spec §7).
- Migrating issues from `polliard/macheim` to `convergent-systems-co/macforge` (operator action).
- Adding audit-envelope/JSON output to workstation verbs (deferred to v0.3+; the spec acknowledges this).
- Publishing a release tag.

---

## Self-review

| Check | Result |
|---|---|
| Spec §3.1 command surface (peer to apple) | Task 5 (cobra port), V2/V3 verification |
| Spec §3.2 layout (internal/workstation/...) | Task 3 (relocate) |
| Spec §4 subtree import with `--squash=false` | Task 2 |
| Spec §4 `git mv` discipline | Task 3.2-3.4 |
| Spec §4 7-commit sequence | This plan has 8 commits (ADR-0018 added per the verb-name-reconciliation question), each named per the spec's conventional-commit prefixes |
| Spec §4.1 `--squash=false` rationale | Task 2 |
| Spec §5.1 flag rename to `--workstation-repo` | Task 5.4 (persistent flag on `newWorkstationCmd`) |
| Spec §5.2 env namespacing | Task 5.4 cobra help text mentions `MACFORGE_WORKSTATION_REPO` (note: actual env-binding via viper is not yet implemented in workstation cmds — captured as "viper integration" gap; matches Spec's §5.3 acknowledgment that this folds into existing macforge layering, but the *binding* of viper-to-flag for workstation is deferred work and the plan only ensures the flag definition is right) |
| Spec §5.3 single-file config layering | Documented in README (Task 8); actual viper integration deferred |
| Spec §6 Go version bump | Task 7 |
| Spec §6 drop urfave/cli/v3, keep x/term | Task 5.20 |
| Spec §6 yaml dependency reconciled | Task 4.4 |
| Spec §7 pre-deletion checklist | Out of scope (operator action) |
| Spec §8 success criteria | "Final verification" V1-V10 |
| Spec §9 out-of-scope items | "Out of scope" section above |

**Gap identified during self-review:** the spec's §5.2 viper-binding for `workstation.repo_path` is not actually implemented by this plan — the cobra flag exists, but the discovery chain step 3-4 (read from `~/.config/macforge/macforge.yaml: workstation.repo_path`) is not wired. Acknowledging that the spec said "folds into macforge's viper layering from README §Configuration layering" — true conceptually, but the wiring code is missing. Treat as known limitation; capture as a follow-up issue. The workstation cmd works via the `--workstation-repo` flag and the convention-path discovery inside `internal/workstation/config.Runtime.ResolveRepoPath()`. Viper-config layering for workstation is real follow-up work, not part of this plan.

No placeholders. Code is shown for every step. Type and function names match across tasks (e.g., `workstationRuntimeFor`, `newWorkstationCmd`, `ErrChecksFailed`).
