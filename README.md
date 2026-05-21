# MacForge

> Civilization-grade Apple release infrastructure for macOS software.

MacForge provides deterministic, auditable, repeatable Apple signing,
certificate lifecycle management, packaging, notarization, verification,
and publishing for macOS software distributed outside the App Store.

[![CI](https://github.com/convergent-systems-co/macforge/actions/workflows/ci.yml/badge.svg)](https://github.com/convergent-systems-co/macforge/actions/workflows/ci.yml)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

## What is this?

MacForge is a single Go binary plus a GitHub Action. It wraps Apple's CLI
toolchain (`codesign`, `security`, `xcrun notarytool`, `spctl`,
`productbuild`) behind one auditable boundary, producing a structured
JSONL audit log for every release.

## Who is it for?

- Maintainers distributing macOS software outside the App Store who want
  a deterministic, scriptable signing pipeline.
- CI/CD platforms that need first-class GitHub Actions / GitLab / Azure
  DevOps integration.
- Enterprises that want a full audit trail without paywalled tooling.

## v0.1 status

v0.1 covers the foundation through sign+verify. See
[`GOALS.md`](GOALS.md) for the full scope and `docs/adr/` for every
load-bearing decision.

| Verb                       | v0.1 status |
|----------------------------|-------------|
| `macforge init`            | ✓           |
| `macforge keychain *`      | ✓           |
| `macforge identity import` | ✓           |
| `macforge identity list`   | ✓           |
| `macforge identity status` | ✓           |
| `macforge sign`            | ✓           |
| `macforge verify`          | ✓           |
| `macforge package`         | stub (v0.2) |
| `macforge notarize`        | stub (v0.2) |
| `macforge publish`         | stub (v0.2) |
| `macforge release`         | stub (v0.2) |

## Quick start

```bash
# 1. Install
go install github.com/convergent-systems-co/macforge/cmd/macforge@latest

# 2. Scaffold project config
cd ~/code/myapp
macforge init --team XYZ1234567

# 3. Create a dedicated keychain
export MACFORGE_KEYCHAIN_PASSWORD=$(openssl rand -base64 24)
macforge keychain create

# 4. Import a Developer ID certificate
macforge identity import --file ./DeveloperID.cer

# 5. Sign your build
macforge sign ./build/MyApp.app

# 6. Verify
macforge verify ./build/MyApp.app
```

## How do I run it?

See [`docs/adr/0005-state-and-config-layout.md`](docs/adr/0005-state-and-config-layout.md)
for the config layout and [`docs/adr/0006-output-format-dual.md`](docs/adr/0006-output-format-dual.md)
for the human/JSON output contract.

## How do I contribute?

See [`CONTRIBUTING.md`](CONTRIBUTING.md). All ADRs live in
[`docs/adr/`](docs/adr/); follow MADR format. Tests follow the three-tier
discipline in [`docs/adr/0013-testing-strategy-three-tiers.md`](docs/adr/0013-testing-strategy-three-tiers.md).

## License

Apache 2.0 — see [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE).
