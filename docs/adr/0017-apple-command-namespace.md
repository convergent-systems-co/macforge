# ADR 0017 — Command Namespace: `macforge apple <verb>` (prep for macheim merge)

| Status      | Accepted                                                                  |
|-------------|---------------------------------------------------------------------------|
| Date        | 2026-05-21                                                                |
| Supersedes  | Parts of [ADR-0004](0004-dependency-stack.md) §5 (verb-tree topology only) |
| Issue       | [#14](https://github.com/convergent-systems-co/macforge/issues/14)        |

## Context

Macheim is being merged into MacForge as a peer subsystem. With today's flat verb tree (`macforge init`, `macforge sign`, `macforge keychain ...`, etc.), every macheim verb would either collide with an existing Apple verb or have to use an unrelated namespace, and the operator surface would lose the "what subsystem am I addressing" signal.

Restructuring while there's exactly one user (the maintainer) and an unreleased binary lets us do a hard cutover without migration noise.

## Decision

All current MacForge verbs move under a top-level `apple` subcommand:

```
macforge
├── apple                            # everything Apple-platform (today's macforge)
│   ├── init
│   ├── identity {create,import,list,status,rotate,export}
│   ├── keychain {create,delete,list,unlock}
│   ├── sign <path>
│   ├── verify <path>
│   ├── package <path>               (stub for v0.2)
│   ├── notarize <path>              (stub for v0.2)
│   ├── publish <path>               (stub for v0.2)
│   ├── release                      (stub for v0.2)
│   └── audit                        (stub for v0.2)
├── macheim                          # peer subtree; populated when macheim merges
└── version
```

`version` stays at the root because it's universal across peers. Subsystem-shared concerns added later (e.g., `macforge config validate`) will land at the root only if they truly span peers; subsystem-specific ones (e.g., `macforge apple config validate`) live inside their peer.

### Audit + JSON envelope updates

- `cliRuntime.command` (the audit `command` field) now reads `apple.<verb>` (e.g., `apple.keychain.create`, `apple.identity.import`). The audit log's `command` is the canonical "what was run" field for the JSONL line.
- Result types' `SchemaName()` reads `macforge.v1.apple.<verb>` (e.g., `macforge.v1.apple.sign`). Downstream `--output json` consumers should dispatch on the new schema string.

Both changes are technically a JSON-output schema change. v0.1-dev has no real consumers; the new schema becomes the v1 baseline. When macheim merges, its envelopes will read `macforge.v1.macheim.<verb>` symmetrically.

### Backwards compatibility

**None.** Hard cutover. Pre-release; one user. Aliases would be noise.

## Consequences

### Positive

- **Clean room for macheim.** Verbs can collide between subsystems and the parser will route correctly.
- **`macforge --help` is now legible at a glance** — two top-level commands instead of eleven, with a clear "Apple-platform stuff lives here" signal.
- **`apple.identity.create` etc. in the audit log** makes it grep-clearer which subsystem ran a command.
- **Subsystem-specific config sections become natural** (`apple.sign.hardened_runtime` instead of `sign.hardened_runtime`) — captured as out-of-scope for this ADR but enabled by it.

### Negative

- **`macforge sign` becomes `macforge apple sign`** — three more characters per invocation. Lived with by every multi-platform tool (`gh repo create`, `kubectl apply -f`, `helm install`).
- **Existing tests and docs that referenced flat verbs must be updated** in lockstep with this ADR. Mechanical edit; covered by the PR.

### Neutral

- The `Use:` strings inside each verb cobra command don't change — only the parent registration does. `init_cmd.go`'s `Use: "init"` is still correct; it's just nested deeper now.

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Flat verbs + macheim-specific prefix only** (`macforge sign` + `macforge macheim-sign`) | Less typing for the dominant Apple path | Asymmetric; bakes "Apple is privileged" into the surface; collision-by-verb stays | Rejected — asymmetry compounds over time |
| Per-subsystem binary (`macforge-apple sign`, `macforge-macheim sign`) | Tight isolation | Doubles the install surface; complicates sharing audit/config/credentials | Rejected — peers SHARE state |
| Profile-based switching (`macforge --profile apple sign`) | Single verb tree | Profile concept is invisible in shell history; conflicts with future per-config profiles | Rejected — wrong layer |
| **Top-level `apple` + `macheim` subtrees** (this ADR) | Symmetric; explicit; pluggable | Three extra chars per Apple invocation | **Chosen** |

## Implementation

- `cmd/macforge/apple_cmd.go` (new): `newAppleCmd()` factory; holds every current sub-command.
- `cmd/macforge/root.go`: registers `newAppleCmd()` + `newVersionCmd()` only.
- Each verb file's `newRuntime(command, ...)` updated from `<verb>` → `apple.<verb>` (e.g., `init` → `apple.init`, `keychain.create` → `apple.keychain.create`).
- Each result type's `SchemaName()` updated from `macforge.v1.<verb>` → `macforge.v1.apple.<verb>`.
- All CLI tests updated to prepend `"apple"` in their `SetArgs` calls.
- README + CHANGELOG entries.

## Future work, NOT in this PR

- `cmd/macforge/macheim_cmd.go` lands when macheim merges; same shape as `apple_cmd.go`.
- Per-subsystem config namespacing (`apple.sign.hardened_runtime` instead of `sign.hardened_runtime`) — pending macheim's config needs.
- `macforge config validate` (issue #13) resumes in the new tree as `macforge apple config validate` (subsystem-specific) or `macforge config validate` (cross-subsystem). Decision deferred until #13 resumes.

## Links

- Issue: [#14](https://github.com/convergent-systems-co/macforge/issues/14)
- Related: [ADR-0004](0004-dependency-stack.md) (cobra), the spec at [`../superpowers/specs/2026-05-21-macforge-architecture-design.md`](../superpowers/specs/2026-05-21-macforge-architecture-design.md) §5.
