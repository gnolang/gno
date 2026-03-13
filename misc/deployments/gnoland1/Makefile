REPO_ROOT := $(shell cd ../../.. && pwd)
GENESIS_SHA256 := 2b1a48bd9a47c1e48f3190754fc7982872c08b347c434048889449c1207117af

.PHONY: help generate generate-txs test clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

generate: genesis.json ## Generate genesis.json and verify checksum
	@echo "Verifying genesis.json checksum..."
	@echo '$(GENESIS_SHA256)  genesis.json' | shasum -a 256 -c
	@echo "OK"

genesis.json: gen-genesis.sh
	./gen-genesis.sh

generate-txs: genesis_txs.jsonl ## Generate genesis transactions only

genesis_txs.jsonl: gen-genesis.sh
	./gen-genesis.sh --txs-only

test: genesis.json ## Verify genesis syntax and boot a node
	@test -s genesis.json || (echo "FAIL: genesis.json is empty" && exit 1)
	@gnogenesis verify -genesis-path genesis.json
	@# Boot a node to validate genesis block execution (not just syntax).
	@echo "Booting node to validate genesis block..."
	@DATA=$$(mktemp -d) && \
	timeout 30 gnoland start \
		-lazy \
		-genesis genesis.json \
		-skip-genesis-sig-verification \
		-data-dir "$$DATA" \
		-genesis-balances-file /dev/null \
		-genesis-txs-file /dev/null \
		2>&1 | grep -q "ABCI Handshake App Info" && \
	rm -rf "$$DATA" && \
	echo "OK: genesis.json is valid (node booted successfully)"

clean: ## Remove generated files
	rm -rf genesis.json packages.gen.txt genesis_txs.jsonl genesis-work
