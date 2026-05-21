# ADR 0010 — Package Naming Reconciliation

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

The bootstrap commit (`f452b57`) created an initial `internal/` layout. `GOALS.md` (added in `6b618ac`) names its internal subsystems separately. The two disagree:

| GOALS.md "Internal Systems" | Bootstrap directory   |
|-----------------------------|-----------------------|
| `apple`                     | `internal/apple`      |
| `keychain`                  | `internal/keychain`   |
| `signing`                   | `internal/signing`    |
| `package`                   | `internal/package`    |
| `notarize`                  | `internal/notary` ⚠   |
| `github`                    | `internal/github`     |
| `ci`                        | (missing) ⚠           |
| `config`                    | `internal/config`     |
| `verification`              | `internal/verify` ⚠   |
| (none)                      | `internal/build` ⚠    |

Plus five packages introduced by the design that aren't in either list: three cross-cutting (`audit`, `errors`, `output`) and two domain (`identity`, distinct from `keychain`; and `ci`, which `GOALS.md` lists but the bootstrap omitted).

Resolving these now — before any implementation lands — is cheap. Later it becomes a rename refactor with import churn.

## Decision

Adopt the following final layout. Treat `GOALS.md` as authoritative on verb-vs-noun choice, but defer to Go idiom where it conflicts (Go internal packages favor concise verbs).

| Final directory     | Status   | Notes                                                                       |
|---------------------|----------|-----------------------------------------------------------------------------|
| `internal/apple/`   | keep     | Unchanged                                                                   |
| `internal/keychain/`| keep     | Unchanged                                                                   |
| `internal/signing/` | keep     | Unchanged                                                                   |
| `internal/package/` | keep     | Unchanged (shadows stdlib `package` but the path, not the import alias, matters) |
| `internal/notarize/`| **renamed** from `internal/notary` | Match `GOALS.md`; verb form aligns with CLI verb `macforge notarize` |
| `internal/github/`  | keep     | Unchanged                                                                   |
| `internal/ci/`      | **added** | CI provider detection + helpers; missing from bootstrap                    |
| `internal/config/`  | keep     | Unchanged                                                                   |
| `internal/verify/`  | keep     | `GOALS.md` says "verification"; Go convention favors the verb form; verb wins for the package, the documentation in `GOALS.md` keeps the noun |
| `internal/release/` | **renamed** from `internal/build` | `build` was ambiguous; `release` matches the `macforge release` CLI verb and the orchestrator role |
| `internal/identity/`| **added** | Cert/key/CSR lifecycle; distinct from `keychain` (storage) — `identity` is the cryptographic identity, `keychain` is the container |
| `internal/audit/`   | **added** | JSONL audit writer + redactor; cross-cutting                                |
| `internal/errors/`  | **added** | Sentinel roots, codes, error envelope; cross-cutting                        |
| `internal/output/`  | **added** | Human vs JSON renderers; cross-cutting                                      |

`GOALS.md` does not need to change — its "Internal Systems" list is a directional sketch, not a strict file list. Where the package names diverge from `GOALS.md` wording, a `README.md` in `internal/` will explain.

## Consequences

### Positive

- **Symmetry between CLI verbs and package names** where it matters (`notarize`, `release`).
- **Explicit cross-cutting packages** — no more shoving audit, errors, output into a grab-bag.
- **`identity` separated from `keychain`** clarifies the data model: an identity (cert + key + chain) is stored in a keychain (container). They have separate lifecycles.
- **`internal/README.md`** documents the layout for new contributors.

### Negative

- **Two directories renamed** — `notary → notarize`, `build → release`. Bootstrap commit needs `git mv`. Trivial because there's no code in those directories yet.
- **`internal/package`** shadows the stdlib word "package". In Go, this is fine for a directory; the import path is unambiguous. But local imports should use `pkgname "github.com/.../internal/package"` to avoid identifier collisions inside files. The implementation plan will note this.

### Neutral

- The four added packages (`ci`, `identity`, `audit`, `errors`, `output`) have no code yet — adding the directories now is just `mkdir`.

## Implementation

1. `git mv internal/notary internal/notarize`
2. `git mv internal/build internal/release`
3. `mkdir internal/{ci,identity,audit,errors,output}`
4. Add `internal/README.md` documenting the layout.
5. Drop `.gitkeep` (or a stub `doc.go`) into each empty directory so git tracks it.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Reconcile now per this ADR** | Cheap (no code yet); avoids future rename churn | Two renames in the bootstrap | **Chosen** |
| Keep bootstrap names; update `GOALS.md` to match | No directory renames | `GOALS.md` is the cleaner source of truth; bending it to fit a bootstrap typo is wrong direction | Rejected |
| Take `GOALS.md` literally everywhere (`verification`, not `verify`) | Maximum consistency | Fights Go idiom; package names should be short verbs where possible | Rejected — idiom matters more than strict consistency |
| Add the missing packages, defer the renames | Less mechanical change now | Pays the rename cost later, when there is code to move | Rejected — every day deferred makes it worse |

## Links

- Spec §4: [Module Organization (final)](../superpowers/specs/2026-05-21-macforge-architecture-design.md#4-module-organization-final)
- Related: [ADR-0002](0002-project-layout-cmd-internal.md)
