# `internal/` — Package Layout

Each subdirectory under `internal/` is a focused package. The `internal/`
prefix means the Go compiler refuses external imports — only the MacForge
binary can use these packages.

## Apple-tool boundary (the one place we shell out)

| Package                  | Responsibility                                        |
|--------------------------|-------------------------------------------------------|
| `apple/`                 | `Runner` interface + ExecRunner + FakeRunner          |
| `apple/codesign/`        | wraps `codesign`                                       |
| `apple/security/`        | wraps `security` (keychain CLI)                        |
| `apple/notarytool/`      | wraps `xcrun notarytool`                               |
| `apple/productbuild/`    | wraps `productbuild` / `pkgbuild`                      |
| `apple/spctl/`           | wraps `spctl`                                          |

## Domain packages

| Package        | Responsibility                                       |
|----------------|------------------------------------------------------|
| `identity/`    | cert/key/CSR lifecycle                               |
| `keychain/`    | dedicated keychain lifecycle                         |
| `signing/`     | sign orchestration                                   |
| `package/`     | zip/dmg/pkg/app (later phases)                       |
| `notarize/`    | submit→wait→staple (later phases)                    |
| `verify/`      | codesign + spctl + Gatekeeper                        |
| `release/`     | end-to-end pipeline orchestration (later phases)     |
| `github/`      | GitHub Releases client (later phases)                |
| `ci/`          | CI provider detection + helpers (later phases)       |

## Cross-cutting

| Package        | Responsibility                                       |
|----------------|------------------------------------------------------|
| `config/`      | `macforge.yaml` + viper layering                     |
| `audit/`       | JSONL audit writer + redactor                        |
| `errors/`      | `MF-*` codes + sentinels + Error envelope (imports as `mferrors`) |
| `output/`      | human vs JSON renderers                              |

See `docs/adr/0010-package-naming-reconciliation.md` for the renaming rationale.
