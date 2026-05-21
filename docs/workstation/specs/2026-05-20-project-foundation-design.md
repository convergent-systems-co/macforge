# Design — Project foundation (sub-epic #4)

**Tracks:** [Index #3](https://github.com/polliard/macheim/issues/3) → [Epic #25 Foundation](https://github.com/polliard/macheim/issues/25) → sub-epic [#4](https://github.com/polliard/macheim/issues/4) → leaves [#30](https://github.com/polliard/macheim/issues/30) [#31](https://github.com/polliard/macheim/issues/31) [#32](https://github.com/polliard/macheim/issues/32) [#33](https://github.com/polliard/macheim/issues/33) [#34](https://github.com/polliard/macheim/issues/34)
**Status:** Draft — awaiting user review
**Date:** 2026-05-20
**Author:** Thomas Polliard (with Claude Opus 4.7)

**Leaf mapping:**
| Spec section | Leaf issue |
|---|---|
| §3 (Go module + tools.go) + §10 (tools.go) | [#30](https://github.com/polliard/macheim/issues/30) Go module + dependency pinning |
| §5 (Makefile) + §7 (.gitignore) + §8 (.editorconfig) | [#31](https://github.com/polliard/macheim/issues/31) Makefile (build/embed-sync/install/test/lint/tidy/clean/help) |
| §6 (golangci.yml) | [#32](https://github.com/polliard/macheim/issues/32) golangci-lint baseline |
| §11 (LICENSE + NOTICE) | [#33](https://github.com/polliard/macheim/issues/33) LICENSE + NOTICE (Apache-2.0) |
| §4 (Layout) + §9 (main.go) + §12 (README) | [#34](https://github.com/polliard/macheim/issues/34) Directory skeleton + main.go + .gitignore + .editorconfig + README placeholder |

---

## 1. Objective

Establish the build / test / lint toolchain and the directory skeleton for `macheim`, such that `make build`, `make test`, and `make lint` are all green on an otherwise empty project. No CLI behavior, no commands, no embedded files yet — just a solid foundation every later sub-issue can build on.

## 2. Rationale

The Epic decomposition makes this the first brick deliberately. Every later sub-issue needs the layout, the lint baseline, the Makefile, and the pinned dependency versions to exist. Landing them here means later issues are pure feature work, not "feature plus tooling drift."

### Alternatives table — binary output location

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| `./macheim` (repo root) | Simplest; `./macheim` runs immediately | Clutters repo root; risk of accidental commit if `.gitignore` slips | Rejected |
| `./bin/macheim` | Conventional Go layout | Conflicts with no actual convention — `cmd/` is the Go convention for source, `bin/` for output is mixed | Rejected |
| `./dist/macheim` | Pairs with future release tooling (GoReleaser); clean root; conventional name for "build artifacts" | One extra path segment to type | **Chosen** |

### Alternatives table — lint strictness

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Minimal (govet + gofmt) | Zero friction; nothing fails on day one | Defers real lint discipline; rusts unused-import etc. before noticed | Rejected |
| Standard (errcheck, govet, ineffassign, staticcheck, gofmt, unused) | Catches the common defects; well-known set; low noise | Skips style and complexity signals | Rejected |
| Standard + extras (above + gocyclo, misspell, revive, gocritic) | Better signal on style and complexity early when codebase is small enough to fix | More tuning during initial sub-issues; some false positives | **Chosen** |

### Alternatives table — license

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| No license now | Repo is private; defer until shareable | Sets a "later" precedent; license is easy to add now and ratchets up trust | Rejected |
| MIT | Maximally permissive; shortest text | No explicit patent grant | Rejected |
| Apache-2.0 | Permissive + explicit patent grant + NOTICE file pattern | Slightly longer header text | **Chosen** |
| BSD-3-Clause | Permissive; concise | No explicit patent grant; less common in Go ecosystem | Rejected |

### Alternatives table — dependency pinning timing

| Alternative | Pros | Cons | Verdict |
|---|---|---|---|
| Pin `urfave/cli/v3` + `yaml.v3` in sub-issue #5 when actually imported | "Don't add what you don't use" | Sub-issue #5 becomes "feature + version selection"; harder to review | Rejected |
| Pin in sub-issue #4 even though unused (this design) | Keeps #5 a pure feature PR; lock-in early | `go mod tidy` will complain unless we `_ "..."` import them | **Chosen** |

The unused-import friction is solved by a `tools.go`-style file that blank-imports them with a `//go:build tools` constraint, so they appear in `go.sum` and the version is locked but they aren't compiled into the binary.

## 3. Scope

### Files to create
- `go.mod` — module path `github.com/polliard/macheim`, Go `1.26`
- `go.sum` — populated by `go mod tidy`
- `main.go` — minimal compilable entry: `package main; func main() {}`
- `Makefile` — targets listed in §5
- `.gitignore` — covers `dist/`, `*.test`, `*.out`, OS junk
- `.golangci.yml` — config from §6
- `.editorconfig` — 2-space YAML/MD, tab Go, trim trailing whitespace
- `LICENSE` — Apache-2.0 with `Copyright 2026 Thomas Polliard`
- `NOTICE` — Apache-2.0 attribution boilerplate
- `tools.go` — blank-import file with `//go:build tools` for dependency lock-in
- `cmd/.gitkeep` — empty
- `internal/config/.gitkeep`
- `internal/embedded/.gitkeep`
- `internal/embedded/scripts/.gitkeep`
- `internal/embedded/configs/.gitkeep`
- `internal/shell/.gitkeep`
- `internal/brew/.gitkeep`
- `internal/dotfiles/.gitkeep`
- `internal/gitrepo/.gitkeep`
- `README.md` — replace the existing 9-byte placeholder with a one-paragraph "what is this" + "see GOALS.md" pointer. Full README is sub-issue #22.

### Files to modify
- `README.md` (currently 9 bytes) → ~10 lines

### Files NOT touched
- `GOALS.md` — already present, immutable
- `.git/` — already configured

## 4. Layout

```
macheim/
├── GOALS.md                  # already present
├── README.md                 # placeholder; full version is #22
├── LICENSE
├── NOTICE
├── Makefile
├── go.mod
├── go.sum
├── main.go
├── tools.go                  # //go:build tools — blank imports for version lock
├── .gitignore
├── .editorconfig
├── .golangci.yml
├── docs/
│   └── superpowers/specs/    # this document lives here
├── cmd/                      # subcommand files arrive in sub-issues #5+
│   └── .gitkeep
└── internal/
    ├── config/.gitkeep
    ├── embedded/
    │   ├── .gitkeep
    │   ├── scripts/.gitkeep
    │   └── configs/.gitkeep
    ├── shell/.gitkeep
    ├── brew/.gitkeep
    ├── dotfiles/.gitkeep
    └── gitrepo/.gitkeep
```

## 5. Makefile

```makefile
SHELL          := /bin/bash
.SHELLFLAGS    := -eu -o pipefail -c
.DEFAULT_GOAL  := help

BINARY         := macheim
DIST_DIR       := dist
BINARY_PATH    := $(DIST_DIR)/$(BINARY)
INSTALL_PREFIX ?= /usr/local
GO             ?= go
GOLANGCI_LINT  ?= golangci-lint

# Build-time identity (overridable from CI)
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
              -X 'main.version=$(VERSION)' \
              -X 'main.commit=$(COMMIT)' \
              -X 'main.buildDate=$(BUILD_DATE)'

.PHONY: help build embed-sync install test lint tidy clean

help:                  ## Show this help
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: embed-sync      ## Build the macheim binary to dist/
	@mkdir -p $(DIST_DIR)
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY_PATH) ./

embed-sync:            ## Sync repo Brewfile/dotfiles into the embedded fallback (no-op if absent)
	@mkdir -p internal/embedded/scripts internal/embedded/configs
	@if [ -f Brewfile ]; then cp Brewfile internal/embedded/configs/Brewfile; fi
	@if [ -d dotfiles ]; then \
	  rm -rf internal/embedded/configs/dotfiles && \
	  cp -R dotfiles internal/embedded/configs/dotfiles; \
	fi

install: build         ## Install to $(INSTALL_PREFIX)/bin
	install -m 0755 $(BINARY_PATH) $(INSTALL_PREFIX)/bin/$(BINARY)

test:                  ## Run unit tests
	$(GO) test -race ./...

lint:                  ## Run golangci-lint
	$(GOLANGCI_LINT) run ./...

tidy:                  ## go mod tidy
	$(GO) mod tidy

clean:                 ## Remove build artifacts
	rm -rf $(DIST_DIR)
```

Notes:
- `embed-sync` is no-op safe: `Brewfile` and `dotfiles/` won't exist yet; both `if` checks short-circuit. Sub-issue #21 lands those sources.
- `VERSION`/`COMMIT`/`BUILD_DATE` are injected via `ldflags` into `main` package variables. `main.go` declares them; sub-issue #5 wires them into `--version`.
- `make help` is the default and self-documents via `## comment` annotations.

## 6. Lint config (`.golangci.yml`)

```yaml
version: "2"

run:
  go: "1.26"
  timeout: 5m
  tests: true

linters:
  enable:
    # standard set
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    # extras
    - gocyclo
    - misspell
    - revive
    - gocritic
  settings:
    gocyclo:
      min-complexity: 12
    revive:
      severity: warning
      rules:
        - name: exported
          arguments: [disableStutteringCheck]
        - name: package-comments
          disabled: true   # we'll add when packages have public surface
    misspell:
      locale: US

formatters:
  enable:
    - gofmt

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

## 7. `.gitignore`

```
# Build artifacts
/dist/

# Test artifacts
*.test
*.out
coverage.out

# OS / editor junk
.DS_Store
*.swp
.idea/
.vscode/

# Optional local override
.envrc
```

## 8. `.editorconfig`

```ini
root = true

[*]
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.go]
indent_style = tab
indent_size = 4

[{*.yml,*.yaml,*.md,Makefile}]
indent_style = space
indent_size = 2

[Makefile]
indent_style = tab
```

## 9. `main.go`

```go
// Package main is the entry point for the macheim CLI.
package main

// Build-time identity, populated via -ldflags. Sub-issue #5 wires these into
// the urfave/cli/v3 root command's --version flag.
var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	// Command tree lands in sub-issue #5.
	_ = version
	_ = commit
	_ = buildDate
}
```

The three `_ = X` reads keep `unused` lint quiet on the package-level variables. They're removed in sub-issue #5 when the variables become real consumers (`--version` flag, build-info subcommand).

## 10. `tools.go` (dependency lock-in)

```go
//go:build tools

// Package main: blank imports here lock dependency versions in go.mod / go.sum
// before they are imported in production code. Removed file in a later sub-issue
// once every listed module is imported from real code.
package main

import (
	_ "github.com/urfave/cli/v3"
	_ "gopkg.in/yaml.v3"
)
```

The `//go:build tools` constraint excludes this file from the default build, so the binary doesn't link them in, but `go mod tidy` keeps them in `go.sum`.

## 11. `LICENSE` and `NOTICE`

Apache-2.0 license text (verbatim from https://www.apache.org/licenses/LICENSE-2.0.txt) plus a `NOTICE` file:

```
macheim
Copyright 2026 Thomas Polliard

This product includes software developed by Thomas Polliard.
```

## 12. `README.md` (placeholder for now)

````markdown
# macheim

> *Old Norse for "Mac home."*

A macOS setup CLI that bootstraps and continuously syncs a Mac to a
known-good state defined in a git repo. See [`GOALS.md`](GOALS.md) for the
full design; full documentation arrives in [issue #22](https://github.com/polliard/macheim/issues/22).

## Status

Foundation skeleton only. No commands implemented yet. See
[Epic #3](https://github.com/polliard/macheim/issues/3) for the delivery
plan.

## Build

```bash
make build
./dist/macheim
```
````

## 13. Testing strategy

Sub-issue #4 doesn't ship behavior to test. But `make test` must pass — meaning `go test ./...` must succeed on a tree with no `_test.go` files. Go's `test` exits 0 with `[no test files]` notices on every package, which is success.

No fixture tests added in this sub-issue. Real tests start in sub-issue #6 (`config` discovery).

## 14. Risk assessment

| Risk | Likelihood | Mitigation |
|---|---|---|
| `golangci-lint` not installed on dev machine, `make lint` fails | Medium | `make lint` runs `golangci-lint`; if absent, prints a clear message and exits non-zero. CI install is in sub-issue #23. README mentions installation. |
| `go.sum` drift when later sub-issues add real imports | Low | `make tidy` is a target; sub-issue #5 runs it. |
| `tools.go` blank-imports break in older Go (no `//go:build` constraint) | Low | Go 1.26 supports `//go:build` natively. |
| Apache-2.0 NOTICE file required by license | Low | NOTICE file included with this sub-issue. |
| `make embed-sync` accidentally copies stale `dotfiles/` over a real one in a future state | Low-Medium | `embed-sync` is the *sync direction* repo→embedded only; never the other way. Comment in Makefile clarifies. |

## 15. Backward compatibility

N/A — first release; no existing surface.

## 16. Dependencies (on other sub-issues)

- **Blocks:** every other sub-issue (#5–#23).
- **Blocked by:** none.

## 17. Out of scope (deferred)

- CLI root command and global flags → **#5**
- Config & repo discovery → **#6**
- Embed system implementation (`//go:embed` directives, resolver) → **#7**
- Shell utilities → **#8**
- Real Brewfile / dotfiles content (sources for `embed-sync`) → **#21**
- README full content → **#22**
- GitHub Actions CI workflow → **#23**

## 18. Acceptance verification

Maps each issue #4 AC to a verifying command:

| AC | Verifying command | Expected |
|---|---|---|
| `go.mod` declares module `github.com/polliard/macheim`, Go 1.26+ | `head -2 go.mod` | `module github.com/polliard/macheim` + `go 1.26` |
| Directory skeleton matches GOALS.md PROJECT LAYOUT | `find . -path ./.git -prune -o -type d -print \| sort` | shows all `internal/*` paths |
| Makefile targets present | `make help` | lists `build embed-sync install test lint tidy clean` |
| `.gitignore` covers Go artifacts | `grep -E '^/dist/|^\*\.test|^\*\.out' .gitignore` | three matches |
| `.golangci.yml` baseline | `golangci-lint linters -c .golangci.yml \| grep -c 'enabled'` | ≥ 9 |
| Deps pinned | `grep -E 'urfave/cli/v3\|yaml.v3' go.mod` | two matches |
| `make build` produces a runnable binary | `make build && ./dist/macheim` | exit 0, no output |
| `make test` passes | `make test` | exit 0 |
| `make lint` passes | `make lint` | exit 0 |

## 19. Commit plan

One PR, one merge commit. Suggested intermediate commits (Code.md §11.2 isolation):
1. `chore: add Apache-2.0 LICENSE and NOTICE`
2. `chore: add .gitignore, .editorconfig`
3. `feat: scaffold Go module + directory skeleton`
4. `feat: add Makefile with build/test/lint/embed-sync targets`
5. `feat: add .golangci.yml lint baseline`
6. `chore: add tools.go to pin urfave/cli/v3 and yaml.v3`
7. `docs: replace README placeholder with foundation notice`

Each commit independently passes `make lint && make test`. Closing line on the final commit:

```
Closes #4
```
