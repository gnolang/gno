REPO_ROOT := $(shell cd ../../.. && pwd)
GENESIS_SHA256 := e6b7db33b1492db875f8cdf8add7d28fe884ec39196459718eba3464cfd43ba0

.PHONY: help generate generate-txs test clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

generate: genesis.json ## Generate genesis.json and verify checksum
	@echo "Verifying genesis.json checksum..."
	@shasum -a 256 genesis.json
	@echo '$(GENESIS_SHA256)  genesis.json' | shasum -a 256 -c

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
