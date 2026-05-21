# ADR 0009 — macforge-action: Separate Repository

| Status | Accepted   |
|--------|------------|
| Date   | 2026-05-21 |

## Context

`GOALS.md` calls out a `macforge-action` GitHub Action with responsibilities: keychain import, signing, packaging, notarization, verification, publishing. The Action is the primary CI consumer of the MacForge CLI.

Two distribution shapes are possible:

1. **Monorepo:** the Action's `action.yml` and its associated scripts live in the same repo as the CLI, perhaps under `action/`.
2. **Separate repo:** `convergent-systems-co/macforge-action` is its own repository, consuming `macforge` as a versioned external binary.

The choice affects release cadence, Marketplace listing, contributor mental model, and upgrade ergonomics for users.

## Decision

**Separate repository:** `convergent-systems-co/macforge-action`.

- The Action repo contains: `action.yml`, a small shim script, examples, README, and its own CI.
- The Action repo pins the `macforge` binary version it consumes (downloaded from a GitHub Release of `convergent-systems-co/macforge` by tag).
- The Action repo has its own SemVer cadence. Major versions are tagged as `v1`, `v2`, … per GitHub Marketplace conventions.
- The Action repo's tests run a real release flow against a test fixture project (uses GitHub's `macos-14` hosted runner).
- The Action's `README.md` is the Marketplace listing; the Action repo is what gets published to Marketplace.

## Consequences

### Positive

- **Independent release cadence.** A patch to the Action shim doesn't require a CLI release, and vice versa.
- **Marketplace expects a separate repo.** GitHub Marketplace publishes from a repository; an `action.yml` in a subdirectory of the CLI repo creates listing ergonomics gotchas (paths, README discovery).
- **Clearer mental model for users.** "I use the CLI" vs. "I use the Action" maps cleanly to two repos.
- **CLI repo stays focused on the CLI.** Fewer issues triaged in the wrong place.
- **Pinned-version model is explicit.** Users see exactly which CLI version the Action runs.

### Negative

- **Two repos to maintain.** Two `CONTRIBUTING.md` files, two issue trackers, two CI surfaces.
- **Cross-repo coordination cost.** A CLI breaking change requires an Action release before users can adopt it. Acceptable — that's also when the matching changelog warning lives.

### Neutral

- The CLI repo's CI publishes the binary; the Action repo's CI verifies the binary works in a real GitHub Actions environment.

## Repository wiring

```
convergent-systems-co/macforge          (this repo — the CLI)
└── publishes signed binaries to GitHub Releases

convergent-systems-co/macforge-action   (sibling repo)
├── action.yml                          composite action: download + run macforge
├── scripts/setup.sh                    fetches the pinned MacForge release
├── examples/
│   └── sign-and-release.yml            end-to-end usage example
├── README.md                           Marketplace listing
└── .github/workflows/test.yml          runs the Action against a fixture project
```

The Action is a **composite action**, not a Docker action — composite actions cache better, are faster to start, and don't impose a container runtime on `macos-*` runners (which don't reliably support Docker).

## Alternatives Considered

| Alternative | Pros | Cons | Verdict |
|-------------|------|------|---------|
| **Separate repo** | Marketplace-native; independent cadence; clear user mental model | Two repos to maintain | **Chosen** |
| Monorepo `action/` subdirectory | One repo to manage; CLI and Action versions move together | Marketplace listing awkward; CLI repo issues get Action issues mixed in; coupled releases create churn | Rejected — friction outweighs convenience |
| Docker action in the CLI repo | Single artifact (a container image) | macOS runners' Docker story is poor; defeats the lightweight composite design | Rejected — wrong tech fit |

## Links

- Spec §3, §11: [Decision Catalog](../superpowers/specs/2026-05-21-macforge-architecture-design.md#3-decision-catalog), [Release Chain](../superpowers/specs/2026-05-21-macforge-architecture-design.md#11-release-chain-bootstrap--self-perpetuating)
- Related: [ADR-0007](0007-release-dogfood-and-bootstrap.md)
