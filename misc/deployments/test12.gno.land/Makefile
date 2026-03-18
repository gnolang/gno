REPO_ROOT := $(shell cd ../../.. && pwd)
GENESIS_SHA256 := 259ce889d28fdfe81659d97c484b880674cabbdc71c48f6ddab260c98ae2342d
GENESIS_NO_AIRDROP_SHA256 := 5e84822840435e3fad2dc00cdcd8220766a392ddf02fc50875b471ba10c2c8c2

# verify_checksum $(file) $(expected_sha256)
define verify_checksum
	@actual=$$(shasum -a 256 $(1) | awk '{print $$1}'); \
	expected=$(2); \
	printf "\nVerify expected $(1) checksum: "; \
	if [ "$$actual" = "$$expected" ]; then \
	printf "MATCH ✅\n  Actual: %s\n" "$$actual"; \
	else \
		printf "MISMATCH ❌\n  Expected: %s\n  Actual:   %s\n" "$$expected" "$$actual"; \
		exit 1; \
	fi
endef

.PHONY: help generate generate-no-airdrop generate-txs test clean

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

generate: ## Generate genesis.json with airdrop balances and verify checksum
	./gen-genesis.sh
	$(call verify_checksum,genesis.json,$(GENESIS_SHA256))

genesis.json: gen-genesis.sh
	./gen-genesis.sh

generate-no-airdrop: ## Generate genesis.json with test1 balance only and verify checksum
	./gen-genesis.sh --no-airdrop
	$(call verify_checksum,genesis.json,$(GENESIS_NO_AIRDROP_SHA256))

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
