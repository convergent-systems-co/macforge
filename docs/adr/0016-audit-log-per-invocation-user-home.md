# ADR 0016 — Audit Log: Per-Invocation JSONL at ~/.macforge/audit/

| Status      | Accepted                                                                  |
|-------------|---------------------------------------------------------------------------|
| Date        | 2026-05-21                                                                |
| Supersedes  | Parts of [ADR-0012](0012-audit-log-schema.md) (rotation, location, project-local mirror) |
| Issue       | [#5](https://github.com/convergent-systems-co/macforge/issues/5)          |

## Context

ADR-0012 placed the audit log at `./.macforge/audit/<UTC>.jsonl` — project-local, daily-rotated. Two problems surfaced during v0.1 testing:

1. **Project-local pollution.** Every `git status` in a repo where MacForge had run showed an untracked `.macforge/` directory.
2. **No cross-project audit view.** An operator wanting to grep "what did I sign last week" had to crawl multiple project directories.
3. **Daily rotation conflates invocations.** A `macforge release` shipping a per-build audit slice has to filter the daily file by `trace` — an extra step that's easy to forget.

The maintainer chose to relocate audit out of projects and into the user home, with one file per `macforge` invocation.

## Decision

The audit log is **per-invocation, user-home, JSONL**:

```
~/.macforge/audit/<UTC-date-time>Z-<trace>.jsonl
```

Filename format: `2006-01-02T15-04-05Z-<26-char-ULID>.jsonl`. The `-` instead of `:` in the time keeps the filename POSIX-safe; lexicographic sort matches chronological order.

- **Per-invocation:** one file per `macforge` run. The Writer takes a filename at construction time; rotation logic is gone.
- **User-home only:** project-local `./.macforge/audit/` is no longer written. Cross-project consolidation in one place.
- **Format:** unchanged — JSONL, schema versioned per ADR-0012 §schema.
- **Extension:** `.jsonl`. The user originally requested `.log`, but `.log` implies freeform text; keeping `.jsonl` preserves editor highlighting, `jq` parseability, and the schema contract.
- **Path root:** `~/.macforge/` rather than `${XDG_STATE_HOME:-$HOME/.local/state}/macforge/`. The XDG-state path is more "correct" per the freedesktop spec; `~/.macforge/` was the maintainer's preference for predictability and brevity. Honors `$HOME`; falls back to relative `.macforge/audit` only when `$HOME` is unset (extremely unusual).

## Resolution order for audit-related state

```
~/.macforge/audit/                  audit logs (this ADR)
~/.config/macforge/macforge.yaml    config (ADR-0015)
./macforge.yaml                     optional project config override (ADR-0015)
~/Library/Keychains/                keychain files (Apple-managed)
```

Two roots in `$HOME` for MacForge state: `~/.config/macforge/` for config, `~/.macforge/` for audit. Intentional — config and runtime-state are conceptually distinct.

## Consequences

### Positive

- **Cross-project audit consolidation.** One `ls -lt ~/.macforge/audit/ | head` shows recent activity across every project.
- **No `.macforge/` in `git status`.** Projects stay clean.
- **Per-invocation files = simpler "release artifact" semantics.** Each `macforge release` produces a single audit file already scoped to that release; no `trace`-filter step.
- **Simpler Writer.** No rotation state machine; the Writer just appends to its assigned file.
- **Sortable filenames.** Lexicographic == chronological.

### Negative

- **Disk accretion.** Many small files (one per `macforge` invocation). A heavy user could have thousands per year. Mitigated by ULID sort + standard rm/find cleanup tooling. A future `macforge audit prune --older-than 90d` is on the table.
- **CI ephemeral home.** GitHub Actions runners have throwaway `$HOME`, so the audit-as-build-artifact story now requires an explicit step like `cp ~/.macforge/audit/$LATEST ./dist/audit.jsonl`. A `macforge release --copy-audit-to <path>` flag will fill that gap when notarize/publish lands.
- **Filename length.** ~50 chars per filename (date + ULID + `.jsonl`). Long but readable.

### Neutral

- The Writer's `Path()` method returns the in-use file so result envelopes can surface the audit location to the operator.
- Schema, redaction, and field vocabulary from ADR-0012 are unchanged.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Per-invocation at `~/.macforge/audit/<UTC>-<trace>.jsonl`** | One file per run; sortable; simple Writer; cross-project view | Many files; long filenames; CI artifact needs explicit copy | **Chosen** |
| Daily-rotated at `~/.macforge/audit/<date>.jsonl` (path-only change) | Fewer files; current rotation logic untouched | Doesn't decouple invocations; `release` still has to trace-filter | Rejected — solves only the project-pollution part |
| Per-invocation files + symlink to `today.jsonl` | Per-invocation isolation + aggregated daily view | Symlinks add OS surface; not implemented | Deferred |
| Project-local AND user-home mirror | Belt-and-suspenders; release artifact stays per-project | 2× disk writes; 2× redaction blast radius | Rejected |
| `.log` extension with text format | Easier to skim by eye | Loses schema contract; `jq` doesn't work; `--output json` consumers can't cross-reference | Rejected — format change without justification |
| `${XDG_STATE_HOME:-$HOME/.local/state}/macforge/audit/` | XDG-canonical | Doesn't match maintainer's stated preference; lots of nested dirs | Rejected — maintainer chose `~/.macforge/` |

## Implementation

- `internal/config/paths.go`: `AuditDir()` returns `$HOME/.macforge/audit`; honors `$HOME`; falls back to `.macforge/audit` when unset. `ProjectAuditDir()` removed.
- `internal/audit/writer.go`: `NewWriter(path string, redactor *Redactor)` — takes a full filename, not a directory. Daily-rotation state machine removed. `Path()` method added so result envelopes can surface the file in use.
- `cmd/macforge/runtime.go`: computes the filename as `filepath.Join(config.AuditDir(), now.Format("2006-01-02T15-04-05Z")+"-"+trace+".jsonl")` at runtime construction.
- Tests retired: `TestWriter_DailyRotation`. New: `TestWriter_AllEventsAppendToSameFile`, `TestWriter_Path`. `TestWriter_ZeroTimeChrononUsesToday` reworked to assert the chronon is patched (since there's no longer a "today's file" to verify).

## Migration

Existing `./.macforge/` directories in user projects are harmless — MacForge just stops writing to them. The user may delete them:

```bash
find . -type d -name .macforge -print -prune
# review, then:
find . -type d -name .macforge -exec rm -rf {} +
```

Old audit files in `~/Library/Application Support/MacForge/audit/` (from the never-shipped audit.user_mirror) are similarly harmless to delete.

## Links

- Supersedes [ADR-0012 §Path, §Daily rotation](0012-audit-log-schema.md). The schema, vocabulary, redaction rules, and "hash-don't-store" defaults from ADR-0012 are unchanged.
- Closes [issue #5](https://github.com/convergent-systems-co/macforge/issues/5).
