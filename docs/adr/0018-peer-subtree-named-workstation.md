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
