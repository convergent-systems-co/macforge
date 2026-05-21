# ADR 0006 — Output Format: Dual (Human + JSON)

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

MacForge serves two audiences from the same command surface:

- **Humans** at a terminal reading colorized output, progress, and summaries.
- **CI systems** parsing structured output to make decisions (e.g., GitHub Actions matrix jobs reading the notarization UUID, downstream scripts gating on verification results).

A tool that only speaks "human" forces CI to scrape stdout — fragile. A tool that only speaks "machine" punishes interactive users. Both audiences are first-class.

## Decision

Every command implements an `Outputter` interface and emits results in **two parallel modes** controlled by the global `--output` flag:

- `--output human` (default) — colorized, with progress and a summary block. ANSI auto-disables under `NO_COLOR` or non-TTY stderr.
- `--output json` — stable, versioned JSON to **stdout**; logs to **stderr** (zerolog JSON writer).

**Auto-detection:** when `GITHUB_ACTIONS=true` and `--output` is unset, mode defaults to `json`. GitHub-Actions-specific workflow commands (`::group::`, `::error::`, `::warning::`) are emitted **in addition** to JSON on stderr to drive the GitHub UI.

### JSON envelope (success)

```json
{
  "ok": true,
  "schema": "macforge.v1.sign",
  "trace": "01HVQK7C8XKR1ZB...",
  "command": "sign",
  "result": {
    "artifact": "./build/MyApp.app",
    "identity_sha1": "<fingerprint>",
    "team": "XYZ1234567",
    "hardened_runtime": true,
    "timestamp": "2026-05-21T14:30:23Z"
  }
}
```

### JSON envelope (failure)

```json
{
  "ok": false,
  "schema": "macforge.v1.error",
  "trace": "01HVQK7C8XKR1ZB...",
  "command": "notarize",
  "code": "MF-NOTARIZE-001",
  "op": "notarize.Submit",
  "message": "Apple rejected the submission: invalid signing identity",
  "hint": "Run `macforge identity status` to confirm the cert is not expired",
  "details": {
    "submission_uuid": "...",
    "log_url": "https://..."
  }
}
```

The `schema` field is the versioned contract for downstream consumers. Per command, per major schema change. v1 schemas are frozen at v1.0.

## Consequences

### Positive

- **CI integration is a flag, not a fork.** No special "ci binary".
- **Stable schemas** mean GitHub Actions can rely on output structure across patch releases.
- **`ok` field** lets downstream consumers branch on success without parsing exit codes.
- **GitHub Actions workflow commands** make the GH UI informative without requiring users to install separate "annotation" actions.

### Negative

- **More code paths to test.** Every command needs unit tests for both renderers.
- **Schema versioning is a maintenance commitment.** Adding fields is fine; renaming or removing requires a `macforge.v2.*` schema and a deprecation window.

### Neutral

- The `Outputter` abstraction sits in `internal/output`. Result types (`SignResult`, `NotarizeResult`, `VerifyResult`) implement it; rendering is a single function dispatch on mode.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Dual: human default + `--output json`** | Both audiences first-class; flag-driven; schema versioned | Two renderers per command | **Chosen** |
| Human only + structured audit log | Single output surface; CI reads JSONL out-of-band | Awkward CI ergonomics; correlating audit with command result requires extra glue | Rejected — too much friction for downstream consumers |
| Triple: human + json + github-actions | Best CI ergonomics for GitHub specifically | Three renderers per command; testing burden up 50% | Rejected — fold GH-Actions enhancements into the JSON mode via stderr workflow commands |
| JSON-first, human as renderer | Maximum determinism; one source of truth | Higher upfront cost; harder to make human output truly idiomatic | Rejected — pragmatically dual renderers are cheaper and clearer |

## Links

- Spec §5, §6: [CLI Surface](../superpowers/specs/2026-05-21-macforge-architecture-design.md#5-cli-surface), [Apple-Tool Boundary](../superpowers/specs/2026-05-21-macforge-architecture-design.md#6-apple-tool-boundary)
- Related: [ADR-0004](0004-dependency-stack.md), [ADR-0011](0011-error-model-and-codes.md)
