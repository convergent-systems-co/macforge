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
| `macforge apple init`            | ✓           |
| `macforge apple keychain *`      | ✓           |
| `macforge apple identity import` | ✓           |
| `macforge apple identity list`   | ✓           |
| `macforge apple identity status` | ✓           |
| `macforge apple sign`            | ✓           |
| `macforge apple verify`          | ✓           |
| `macforge apple package`         | stub (v0.2) |
| `macforge apple notarize`        | stub (v0.2) |
| `macforge apple publish`         | stub (v0.2) |
| `macforge apple release`         | stub (v0.2) |

## Quick start

```bash
# 1. Install
go install github.com/convergent-systems-co/macforge/cmd/macforge@latest

# 2. Scaffold the global config (one-time, identity-shaped fields only)
macforge apple init --team XYZ1234567
# Writes ~/.config/macforge/macforge.yaml

# 3. Create a dedicated keychain
export MACFORGE_KEYCHAIN_PASSWORD=$(openssl rand -base64 24)
macforge apple keychain create

# 4. Import a Developer ID certificate
macforge apple identity import --file ./DeveloperID.cer

# 5. (Optional) Per-project overrides — only if your project needs them
cd ~/code/myapp
cat > macforge.yaml <<'YAML'
sign:
  entitlements: ./Entitlements.plist
package:
  formats: [zip, dmg]
publish:
  github:
    repo: convergent-systems-co/myapp
YAML

# 6. Sign your build
macforge apple sign ./build/MyApp.app

# 7. Verify
macforge apple verify ./build/MyApp.app
```

## Configuration layering

MacForge reads config in priority order (highest first):

1. CLI flag — e.g., `--team-id`, `--entitlements`, `--config <path>`
2. Environment — `MACFORGE_*` (e.g., `MACFORGE_KEYCHAIN_PASSWORD`)
3. `./macforge.yaml` — project-local override (optional)
4. `~/.config/macforge/macforge.yaml` — global base (required; created by `macforge apple init`)
5. Built-in defaults

The global file holds identity-shaped fields (team, keychain, signing identity, ASC profile). The project-local file is optional and should carry only project-shaped fields (entitlements, package formats, publish target). See [ADR-0015](docs/adr/0015-single-global-config-xdg.md) for the field-by-field classification.

## How do I run it?

See [ADR-0015](docs/adr/0015-single-global-config-xdg.md) for the config
layout, [ADR-0005](docs/adr/0005-state-and-config-layout.md) for the
project-local audit log location, and
[ADR-0006](docs/adr/0006-output-format-dual.md) for the human/JSON output
contract.

## How do I contribute?

See [`CONTRIBUTING.md`](CONTRIBUTING.md). All ADRs live in
[`docs/adr/`](docs/adr/); follow MADR format. Tests follow the three-tier
discipline in [`docs/adr/0013-testing-strategy-three-tiers.md`](docs/adr/0013-testing-strategy-three-tiers.md).

## License

Apache 2.0 — see [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE).
