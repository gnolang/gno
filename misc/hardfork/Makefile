GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(GOPATH)/bin

.PHONY: help install build test clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

install: ## Install hardfork binary to GOBIN
	go install .

build: ## Build hardfork binary in current directory
	go build -o hardfork .

test: ## Run unit tests
	go test ./...

clean: ## Remove build artifacts
	rm -f hardfork

# Quick local smoke test: generate a genesis from a skip-txs preview and test it.
# Requires a running gnoland RPC or set SOURCE to a local dir/file.
SOURCE ?= http://rpc.gno.land:26657
CHAIN_ID ?= gnoland-1
ORIGINAL_CHAIN_ID ?= gnoland1

preview-and-test: build ## Generate a skip-txs genesis preview and smoke-test it locally
	@echo "==> Generating skip-txs genesis preview from $(SOURCE)..."
	./hardfork genesis \
		--source "$(SOURCE)" \
		--chain-id "$(CHAIN_ID)" \
		--original-chain-id "$(ORIGINAL_CHAIN_ID)" \
		--skip-txs \
		--output /tmp/hardfork-preview.json
	@echo ""
	@echo "==> Testing genesis replay (genesis-mode txs only)..."
	./hardfork test --genesis /tmp/hardfork-preview.json
