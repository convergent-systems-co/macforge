# HANDOFF — MacForge v0.1 implementation

**Last updated:** 2026-05-21 (Phase A complete; Phases B + C pending)

## Where we are

- **Branch:** `v0.1-impl`
- **Worktree:** `/Users/itsfwcp/workspace/convergent-system-co/macforge/.git/worktrees/v0.1-impl`
- **Last commit on branch:** `cb5b763` (`feat(apple/security): wrap security CLI (keychain + identity)`)
- **Commits ahead of main:** 14 (all on `v0.1-impl`; nothing pushed to remote)
- **Working tree:** clean

## What Phase A delivered (Tasks 1–11 of `docs/superpowers/plans/2026-05-21-v0.1-vertical-slice.md`)

Foundation packages and infrastructure — buildable, fully tested, but does not yet sign anything:

| Task | What landed                                                        | Last commit |
|------|--------------------------------------------------------------------|-------------|
| 1    | Naming reconciliation; doc.go stubs; `internal/README.md`          | `a47ffa9`   |
| 2    | LICENSE (Apache 2.0), NOTICE, `.gitignore`                         | `7e6f4c9`   |
| 3    | `internal/errors` (`mferrors`) — MF-* codes, sentinels, Error type | `751ff16`   |
| 4    | `internal/audit` — JSONL writer, redactor, ULID trace ID           | `58a32db`   |
| 5    | `internal/output` — dual-mode renderer (human + JSON envelope)     | `00aad95`   |
| 6    | `internal/config` — Config schema + viper layering + validation    | `6aeacb7`   |
| 7    | `internal/apple` — Runner interface, ExecRunner, FakeRunner        | `ab14e7e`   |
| 8    | `cmd/macforge` — cobra root + global flags + 10 verb stubs         | `533cace`   |
| 9    | CI workflows (lint, unit, integration, vulncheck) + golangci-lint  | `e40f58b`   |
| 10   | `cliRuntime` + `macforge init` (first real verb, end-to-end)       | `d93481d`   |
| 11   | `internal/apple/security` — typed wrapper over `security` CLI       | `cb5b763`   |

## Verifiable state (re-check on resume per Common.md U10)

Before doing anything in Phase B, run these from the worktree root and confirm they match the assertions:

```bash
cd /Users/itsfwcp/workspace/convergent-system-co/macforge/.git/worktrees/v0.1-impl
git status                                # expect: clean, on v0.1-impl
git log --oneline | head -1               # expect: cb5b763 feat(apple/security): ...
go build ./...                             # expect: exits 0
go test -race ./...                        # expect: all packages green
go test -race -tags=integration ./...      # expect: all packages green
```

If any of these fail, do NOT proceed — investigate first.

## What's next: Phase B (Tasks 12–17)

The plan at `docs/superpowers/plans/2026-05-21-v0.1-vertical-slice.md` defines these tasks verbatim with full source code. Continue with subagent-driven execution per the same pattern Phase A used.

| Task | Scope                                                              |
|------|--------------------------------------------------------------------|
| 12   | `internal/keychain` — Manager, naming validator, secret resolver   |
| 13   | `internal/identity` — Service (Import, List, ReadCertStatus)       |
| 14   | CLI verbs: `macforge keychain {create, delete, list, unlock}`      |
| 15   | CLI verbs: `macforge identity {import, list, status}` (stubs for create/rotate/export) |
| 16   | Tier 3 e2e — keychain + identity round-trip on real macOS          |

## What's after: Phase C (Tasks 18–23)

| Task | Scope                                                              |
|------|--------------------------------------------------------------------|
| 17   | `internal/apple/codesign` — Sign/Verify/Display wrapper             |
| 18   | `internal/apple/spctl` — Assess wrapper                             |
| 19   | `internal/signing` — Service that orchestrates codesign + identity |
| 20   | `internal/verify` — Service composing codesign + spctl              |
| 21   | CLI verbs: `macforge sign`, `macforge verify`                       |
| 22   | Tier 3 e2e — sign + verify with HelloApp.app fixture                |
| 23   | README + CONTRIBUTING + SECURITY + CHANGELOG + tag prep             |

## Notable items the subagents touched up vs. the plan

These were caught in review and fixed in follow-up commits — record so re-readers don't think the plan is wrong:

1. **`a47ffa9`** — `internal/errors/doc.go` godoc comment must lead with `// Package mferrors` (declared package name).
2. **`58a32db`** — `audit.Writer.Write` must patch zero-time `Chronon` BEFORE the rotation check, otherwise a zero-time Event opens `0001-01-01.jsonl`.
3. **`ab14e7e`** — On darwin, `exec.CommandContext` killed by timeout DOES return an `*exec.ExitError` (the plan's note said otherwise). `ExecRunner.Run` guards with `ctx.Err() != nil` before the ExitError cast so timeouts route to `mferrors.NewTool(CodeToolMissing, ...)`.
4. **`a9b51c6`** — The bootstrap's empty `internal/*` directories don't exist in this worktree (git doesn't track empty dirs). Task 1's `git mv` steps were no-ops; the renames were achieved by creating only the new layout. Subsequent tasks created the rest of the `internal/` tree as they added files.

## Open assertions for Phase B start

- The `internal/keychain` directory does **not** exist in this worktree yet. Task 12 will create it.
- Same for `internal/signing`, `internal/notarize`, `internal/release`, `internal/verify`, `internal/package`, `internal/github`, `internal/apple/codesign`, `internal/apple/spctl`. Each task creates its own package directory on first write.
- `internal/identity/doc.go` exists with `package identity`; Task 13 will add `identity.go`, `import.go`, `status.go` to that directory.

## Dependencies installed

`go.mod` (verified at commit `cb5b763`):

```
go 1.26.3
github.com/oklog/ulid/v2 v2.1.1            # audit trace IDs
github.com/spf13/cobra v1.10.2             # CLI framework
github.com/spf13/viper v1.21.0             # config layering
gopkg.in/yaml.v3                            # config file parsing (transitive via viper)
```

No further deps planned for Phase B/C unless an unexpected need arises.

## Continuing the work

1. Re-open the worktree (per Common.md U17): `cd /Users/itsfwcp/workspace/convergent-system-co/macforge/.git/worktrees/v0.1-impl`.
2. Confirm the verifiable state block above.
3. Open `docs/superpowers/plans/2026-05-21-v0.1-vertical-slice.md`. Tasks 12–23 are verbatim there.
4. Invoke `superpowers:subagent-driven-development` to continue the same execution pattern Phase A used.
5. Each task: dispatch implementer subagent → spec review (or condensed inspection for verbatim tasks) → code review → fix loop → mark complete.

## Worktree cleanup (only when v0.1 is fully merged)

Per Common.md U17.3: `git worktree remove .git/worktrees/v0.1-impl` (do NOT `rm -rf`).
