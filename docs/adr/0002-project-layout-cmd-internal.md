# ADR 0002 — Project Layout: `cmd/` + `internal/`

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

A Go project of MacForge's scope needs an opinionated layout that signals intent to contributors and prevents accidental external coupling. The Go community has two common conventions: (a) the [`golang-standards/project-layout`](https://github.com/golang-standards/project-layout) pattern with `cmd/`, `internal/`, `pkg/`, `api/`, and friends; (b) a flatter root-package layout used by smaller tools.

MacForge has multiple internal subsystems (`apple`, `signing`, `notarize`, `verify`, …) that need clear boundaries, plus exactly one binary entrypoint. Public re-use of internal packages by third parties is **not** a v0.x goal — and is arguably an anti-goal for a tool whose contract is its CLI.

## Decision

- Single binary entrypoint at `cmd/macforge/main.go`.
- All implementation under `internal/`, organized by subsystem.
- **No `pkg/` directory.** No public Go API surface in v0.x. If external Go consumers are ever wanted, that becomes its own ADR.
- ADRs live under `docs/adr/`. Design specs live under `docs/superpowers/specs/`.
- Examples live under `examples/`. Tier-2 fixtures live under `testdata/`. Tier-3 e2e tests live under `e2e/`.

## Consequences

### Positive

- `internal/` is enforced by the Go compiler — no third party can import these paths, freeing refactoring.
- One binary = one mental model. The CLI is the contract.
- Aligns with `kubectl`, `gh`, `terraform`, `helm`, `hugo` — the layout most Go ops engineers recognize.

### Negative

- If we ever want to expose programmatic Go consumers (e.g., a library), we need to lift code out of `internal/`. That's a deliberate cost — and a deliberate decision moment.

### Neutral

- The `cmd/macforge/` directory has exactly one Go file (`main.go`) at the start; it may grow to a small handful of wiring files. That is fine.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **`cmd/` + `internal/`** | Idiomatic; enforced boundary; familiar | Slight extra path nesting | **Chosen** |
| Flat root package | Less ceremony for tiny projects | Doesn't scale; encourages accidental cross-package leakage; not idiomatic at this size | Rejected — MacForge is not tiny |
| `cmd/` + `pkg/` + `internal/` | Reserves a place for a future public API | Premature; no v0.x consumer; YAGNI | Rejected — add `pkg/` only when there is a real external consumer |
| Hexagonal / DDD-style (`domain/`, `adapters/`, `ports/`) | Cleaner architectural metaphor for some readers | Unfamiliar to most Go contributors; over-architected for the problem | Rejected — premature abstraction (see `Code.md` discipline) |

## Links

- Spec: [`../superpowers/specs/2026-05-21-macforge-architecture-design.md`](../superpowers/specs/2026-05-21-macforge-architecture-design.md)
- Related: [ADR-0001](0001-language-go.md), [ADR-0010](0010-package-naming-reconciliation.md)
