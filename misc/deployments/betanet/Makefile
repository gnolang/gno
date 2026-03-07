REPO_ROOT := $(shell cd ../../.. && pwd)

.PHONY: generate test clean

generate: genesis.json

genesis.json: packages.yml gen-genesis.sh
	@# Ensure gno and gnogenesis are available.
	cd $(REPO_ROOT)/gnovm && go install ./cmd/gno
	cd $(REPO_ROOT)/contribs/gnogenesis && go install .
	./gen-genesis.sh

test: genesis.json
	@test -s genesis.json || (echo "FAIL: genesis.json is empty" && exit 1)
	@gnogenesis verify --genesis-path genesis.json
	@echo "OK: genesis.json is valid"

clean:
	rm -f genesis.json
