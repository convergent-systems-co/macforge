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

# embed-sync is repo -> embedded only. It never copies from embedded
# back to the repo. Sources (Brewfile, dotfiles/) arrive in sub-issue #21;
# until they exist this target is a no-op by design. Idempotent: a
# second invocation with no source changes produces no on-disk diff
# (cmp -s short-circuits Brewfile; rsync -a skips files already up to
# date and --delete keeps the dotfiles tree in exact lockstep).
embed-sync:            ## Sync repo Brewfile/dotfiles into the embedded fallback (idempotent; no-op if absent)
	@mkdir -p internal/embedded/scripts internal/embedded/configs
	@if [ -f Brewfile ]; then \
	  cmp -s Brewfile internal/embedded/configs/Brewfile || cp Brewfile internal/embedded/configs/Brewfile; \
	fi
	@if [ -d dotfiles ]; then \
	  rsync -a --delete dotfiles/ internal/embedded/configs/dotfiles/; \
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
