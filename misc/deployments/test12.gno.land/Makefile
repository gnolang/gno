REPO_ROOT := $(shell cd ../../.. && pwd)
GENESIS_SHA256 := 34586634ed794c5167fbf55a06d8806fcbb8fe558a26399fa8b0e8e8b377b355
GENESIS_NO_AIRDROP_SHA256 := aebbfc9b3af201a287c8fe0de033beb830f1da0520ec2ba37d59a15e45b2accd

.PHONY: help generate generate-no-airdrop generate-txs test clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

generate: genesis.json ## Generate genesis.json with airdrop balances and verify checksum
	@echo "Verifying genesis.json checksum..."
	@shasum -a 256 genesis.json
	@echo '$(GENESIS_SHA256)  genesis.json' | shasum -a 256 -c

genesis.json: gen-genesis.sh
	./gen-genesis.sh

generate-no-airdrop: ## Generate genesis.json with test1 balance only and verify checksum
	./gen-genesis.sh --no-airdrop
	@echo "Verifying genesis.json checksum..."
	@shasum -a 256 genesis.json
	@echo '$(GENESIS_NO_AIRDROP_SHA256)  genesis.json' | shasum -a 256 -c

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
