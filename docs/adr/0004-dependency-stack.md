# ADR 0004 — Dependency Stack: cobra + viper + zerolog

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

MacForge needs:

- A multi-level command tree (`identity create`, `identity import`, `identity rotate`, …) with inheritable flags and generated help/completions.
- Layered configuration (CLI flag → env → project YAML → user YAML → defaults).
- Structured logging that supports both human (console) and JSON output, with redaction hooks.
- An error model that integrates cleanly with the above.

The Go ecosystem has well-trodden paths for all three.

## Decision

- **CLI framework:** [`spf13/cobra`](https://github.com/spf13/cobra).
- **Configuration:** [`spf13/viper`](https://github.com/spf13/viper).
- **Logging:** [`rs/zerolog`](https://github.com/rs/zerolog) — chosen over `uber-go/zap` for smaller surface, zero-allocation hot path, and simpler API. Either is fine; zerolog wins on minimalism.
- **YAML:** [`gopkg.in/yaml.v3`](https://github.com/go-yaml/yaml) (transitive via viper anyway).
- **Errors:** stdlib `errors` package only (`Is`, `As`, `fmt.Errorf("...: %w", err)`) — see ADR-0011.

These come with their transitive dependencies; that's the cost of ecosystem familiarity.

## Consequences

### Positive

- **Discoverable to any Go ops engineer.** `kubectl`, `gh`, `helm`, `hugo`, `argocd` all use cobra. Onboarding cost approaches zero.
- **Completions, help, validation come for free** with cobra.
- **viper handles the layering correctly**, including env-var binding, type coercion, and watch.
- **zerolog is fast and small.** Console writer for human output, JSON writer for `--output json`. Redaction via a custom `Hook`.

### Negative

- **Larger dependency surface than pure stdlib.** Per `Code.md §6`, each dep must be justified — done here. Each must be pinned via `go.mod` + `go.sum`.
- **viper is somewhat opinionated** about config precedence; we may need to override its default merge behavior in places.
- **zerolog's API is unusual** (`.Str("k", v).Msg("...")` chained syntax). Contributors learn it once.

### Neutral

- Lockfile (`go.sum`) is mandatory per `Code.md §6`. Vulnerability scanning (`govulncheck`) runs in CI.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **cobra + viper + zerolog** | Ecosystem-standard for first two; zerolog is the lighter of the two leading log libs | Larger surface than stdlib | **Chosen** |
| cobra + stdlib `slog` + hand-rolled config | Minimal logging dep, no viper opinion | Loses viper's env+layering machinery (must hand-roll); slog handler ecosystem younger | Rejected — viper's value > slog's freshness for this tool |
| Pure stdlib (`flag`, `slog`, `encoding/json`) | Maximum durability, zero third-party rot | No nested subcommands without writing your own dispatcher; no completions; config layering hand-rolled | Rejected — CLI ergonomics matter for adoption |
| cobra + viper + zap | zap has more features and bigger community | More code, more allocation, more API surface; nothing zerolog can't do here | Rejected — zerolog wins on minimalism for this use case |

## Logging conventions (zerolog)

- Console writer to **stderr** in human mode. JSON writer to **stderr** in JSON mode (stdout reserved for command result).
- Every log line carries `trace` (the run ULID) and `subsystem`.
- Redaction is a `zerolog.Hook` that masks substrings declared in `apple.Invocation.Redact`.
- Default level: `info`. `-v` → `debug`. `--log-level trace` enables stdout/stderr-body capture in the audit log (`--audit-bodies`).

## Links

- Related: [ADR-0006](0006-output-format-dual.md), [ADR-0011](0011-error-model-and-codes.md), [ADR-0012](0012-audit-log-schema.md)
- Reference: `Code.md §6 Dependency and Supply-Chain Hygiene`
