# Contributing to MacForge

Thanks for your interest. MacForge is "civilization-grade" infrastructure —
durability and auditability matter more than velocity.

## Before you start

- Read [`GOALS.md`](GOALS.md).
- Skim [`docs/adr/`](docs/adr/) — every load-bearing decision lives there.
- Skim [`docs/superpowers/specs/`](docs/superpowers/specs/) for the
  architecture overview.

## Development setup

```bash
git clone https://github.com/convergent-systems-co/macforge
cd macforge
go test ./...
```

You can develop on Linux/Windows/macOS. Tier 3 e2e tests need a Mac;
the CI runs them on `macos-14`.

## Pull requests

1. **One logical change per commit.** Refactors, features, and bug fixes
   stay separate (Code.md §11.2).
2. **Tests first.** Bug fixes start with a failing regression test
   (Code.md §11.4).
3. **Conventional commit messages.** `feat(scope):`, `fix(scope):`,
   `refactor(scope):`, `docs(scope):`, `chore:`.
4. **Reference an ADR** when your change implements or modifies a
   load-bearing decision.

## ADRs

If your change makes a decision worth remembering — language, dependency,
data layout, error contract, boundary — write an ADR in
[`docs/adr/`](docs/adr/) using MADR format. Number it after the last
existing ADR.

## Code style

See `.golangci.yml`. The CI lint job is the source of truth.

## Security

See [`SECURITY.md`](SECURITY.md). No secrets in artifacts, ever.
