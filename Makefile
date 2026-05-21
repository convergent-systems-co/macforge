# MacForge — convenience wrapper around `go`.
#
# Run `make` (or `make help`) to list targets.

.PHONY: help clean build test embed-sync

BINARY   := macforge
PKG      := ./cmd/macforge
BIN_DIR  := bin
VERSION  := v0.1.0-dev+$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
LDFLAGS  := -X main.version=$(VERSION)
GOFLAGS  := -trimpath

.DEFAULT_GOAL := help

help:  ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[1m%-12s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

clean:  ## Remove build artifacts (./bin, *.test, coverage.out)
	rm -rf $(BIN_DIR) coverage.out
	@find . -name '*.test' -not -path './.git/*' -delete 2>/dev/null || true

build: embed-sync  ## Build ./bin/macforge with the current git-SHA version stamp
	@mkdir -p $(BIN_DIR)
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(PKG)
	@echo "Built: $(BIN_DIR)/$(BINARY) ($(VERSION))"

test:  ## Run unit + integration tests with -race
	go test -race ./...
	go test -race -tags=integration ./...

embed-sync:  ## Sync examples/workstation/Brewfile -> internal/workstation/embedded/configs/Brewfile
	@mkdir -p internal/workstation/embedded/configs
	@cp examples/workstation/Brewfile internal/workstation/embedded/configs/Brewfile
	@echo "embed-sync: examples/workstation/Brewfile -> internal/workstation/embedded/configs/Brewfile"
