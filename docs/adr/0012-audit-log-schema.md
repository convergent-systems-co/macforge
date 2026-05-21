# ADR 0012 — Audit Log: JSONL with `Common.md §5.2` Vocabulary

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

`GOALS.md` mandates "Auditable" and "Full audit trail." Every Apple invocation, every decision, every state transition is interesting to one of three audiences:

- **Operators** debugging a failed release.
- **Security auditors** reconstructing what was signed, by whom, with which identity.
- **Future automation** (compliance reports, SLSA provenance generation).

A flat, line-delimited, append-only JSON log serves all three: greppable on a terminal, parseable in any language, replayable in time order.

A consistent vocabulary across MacForge logs and the user's broader AI/governance audit streams (`~/.ai/audit/interactions/<YYYY-MM>.jsonl`) means one set of grep patterns works everywhere. `Common.md §5.2` already defines that vocabulary deliberately to avoid mirroring any one tool's internal terms.

## Decision

Audit logs are **JSONL**, one event per line, append-only, daily-rotated by UTC date, written to:

- **`./.macforge/audit/<UTC-ISO-8601-date>.jsonl`** (project scope) — always on.
- **`~/Library/Application Support/MacForge/audit/<UTC>.jsonl`** (user scope) — opt-in via config `audit.user_mirror: true`.

### Field vocabulary

Mirrors `Common.md §5.2` deliberately.

| Field            | Type              | Purpose                                                              |
|------------------|-------------------|----------------------------------------------------------------------|
| `chronon`        | RFC3339 UTC, ms   | Event timestamp                                                       |
| `trace`          | ULID              | Per-`macforge` invocation ID; threads through every subcommand        |
| `cwd`            | string            | Working directory at event time                                       |
| `actor`          | enum              | `macforge` \| `user` \| `tool` \| `system`                            |
| `kind`           | enum              | `invocation-attempt`, `invocation-result`, `signal`, `decision`, `error` |
| `probe`          | string            | Apple tool name on invocation events; subsystem name on `signal`      |
| `probe_payload`  | string, ≤1000ch   | argv string on `invocation-attempt`; truncated and redacted           |
| `exit`           | int               | Exit code on `invocation-result`                                      |
| `duration_ms`    | int               | Wallclock duration on `invocation-result`                             |
| `stdout_sha256`  | hex               | sha256 of stdout (body stored separately if `--audit-bodies`)         |
| `stderr_sha256`  | hex               | sha256 of stderr                                                       |
| `artifact_sha256`| hex               | sha256 of the artifact being operated on, when applicable             |
| `redacted`       | array<string>     | Tags describing what was masked (e.g., `["asc-password"]`)            |
| `note`           | string            | Free-form, on `signal` events                                          |
| `code`           | string            | `MF-*` code, on `error` events                                         |
| `op`             | string            | Caller op, on `error` events                                           |
| `team`           | string            | Apple Team ID, when relevant                                           |
| `identity_sha1`  | hex               | Signing-identity fingerprint, when relevant                            |

Field set per event kind:

```
invocation-attempt : chronon, trace, cwd, actor, kind, probe, probe_payload, redacted
invocation-result  : chronon, trace, cwd, actor, kind, probe, exit, duration_ms,
                     stdout_sha256, stderr_sha256, artifact_sha256?, redacted
signal             : chronon, trace, cwd, actor, kind, probe, note, ...event-specific
decision           : chronon, trace, cwd, actor, kind, note, ...decision-specific
error              : chronon, trace, cwd, actor, kind, code, op, note
```

### Rules

| Rule                | Behavior                                                                                  |
|---------------------|-------------------------------------------------------------------------------------------|
| Append-only         | Open with `O_APPEND \| O_CREATE`. Never rewrite an existing line.                          |
| Daily rotation      | Filename is the UTC date of the first event in that file. Cross-midnight runs span files. |
| Redaction first     | `Invocation.Redact` substrings replaced with `[REDACTED:<kind>]` **before** the line is serialized. P4 non-negotiable. |
| Hash, don't store   | stdout/stderr bodies are stored as `sha256` digests by default. `--audit-bodies` opts into storing full bodies under `./.macforge/audit/bodies/<sha256>` (still subject to redaction). |
| Schema versioning   | Top-level `schema: "macforge.audit.v1"` on the first event of each file.                  |
| Trace               | One ULID generated at process start; passed through every Runner call; logged in every line. |
| Determinism         | JSON keys serialized in stable insertion order. No trailing whitespace. UTF-8. LF line endings. |

### What does NOT go in the audit log

- Raw secret values (passwords, ASC API keys, JWT tokens, signing material).
- PII unless explicitly relevant (no email harvesting).
- Source code being signed (only its hash).

If the writer detects that an unredacted secret would land in `probe_payload`, the line is dropped, an `error`-kind audit entry is written instead, and the writer panics in development builds (catches missing `Redact` declarations early).

## Consequences

### Positive

- **Vocabulary shared with `~/.ai/audit/interactions/`** — one grep, two truths.
- **Greppable on a terminal**, parseable anywhere.
- **Append-only + daily rotation** prevents accidental corruption.
- **Hash-by-default** keeps the log small. Bodies on demand for debugging.
- **Schema versioned** at the file level — readers can dispatch on version.

### Negative

- **`--audit-bodies` mode produces lots of data.** Disabled by default; explicitly opt-in.
- **The audit log itself can leak signal.** A path being signed reveals the artifact name. Acceptable — the audit log is for trusted readers, not adversaries.

### Neutral

- The audit log is a **first-class release artifact**. On disk, entries are written to daily-rotated `./.macforge/audit/<UTC-date>.jsonl`. At publish time, `macforge release` filters by the current run's `trace` and consolidates those entries into a single file named `macforge-audit-<version>.jsonl`, which is uploaded alongside the signed binary by default. Disable via `--no-audit-publish`.

## Reader tooling

`macforge audit` provides:

```
macforge audit tail                stream the current day's audit log
macforge audit query --code MF-*  filter by error code
macforge audit query --trace ID   show all events for a trace
macforge audit verify <file>      check schema version + line integrity
```

These are convenience over `jq` — every query is a documented `jq` filter under the hood.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **JSONL with `Common.md §5.2` vocabulary** | Greppable, parseable, append-friendly, cross-tool vocabulary | New vocabulary for first-time MacForge readers | **Chosen** |
| Single growing `audit.log` (no rotation) | Simpler | Unbounded growth; harder to publish per release | Rejected — daily rotation costs nothing |
| SQLite | Queryable, transactional | Binary format; harder to grep / version / publish; tooling overhead | Rejected — wrong shape for an append-only audit |
| OpenTelemetry traces / OTLP | Industry standard for distributed tracing | Wrong abstraction (MacForge is local, not distributed); collector overhead; not auditor-friendly | Rejected — over-engineered for this use case |
| Plain text human log | Easiest to read by eye | Not parseable; can't drive automation | Rejected — defeats the "auditable for automation" goal |

## Links

- Spec §8: [Audit Log](../superpowers/specs/2026-05-21-macforge-architecture-design.md#8-audit-log)
- Reference: `~/.ai/Common.md §5.2` (governance interaction audit log vocabulary)
- Related: [ADR-0003](0003-apple-tool-boundary-shell-out.md), [ADR-0005](0005-state-and-config-layout.md), [ADR-0011](0011-error-model-and-codes.md)
