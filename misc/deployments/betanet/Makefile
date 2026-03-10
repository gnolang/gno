REPO_ROOT := $(shell cd ../../.. && pwd)

.PHONY: generate test clean

generate: genesis.json

genesis.json: gen-genesis.sh
	@# Ensure gno, gnoland, and gnogenesis are available.
	cd $(REPO_ROOT)/gnovm && go install ./cmd/gno
	cd $(REPO_ROOT)/gno.land && go install ./cmd/gnoland
	cd $(REPO_ROOT)/contribs/gnogenesis && go install .
	./gen-genesis.sh

test: genesis.json
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

clean:
	rm -f genesis.json packages.gen.txt
