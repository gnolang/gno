# misc/hf-glue — HIGHLY EXPERIMENTAL hardfork testbed
#
# See README.md. Do not merge.

SHELL := bash

# ---- configurable ----------------------------------------------------------
SOURCE             ?= https://rpc.betanet.testnets.gno.land
ORIGINAL_CHAIN_ID  ?= gnoland1
CHAIN_ID           ?= gnoland-1
HALT_HEIGHT        ?=
VALIDATOR_NAME     ?= hf-glue-local

export SOURCE ORIGINAL_CHAIN_ID CHAIN_ID HALT_HEIGHT VALIDATOR_NAME

# ---- paths -----------------------------------------------------------------
HERE := $(abspath .)
REPO := $(abspath ../..)
OUT  := $(HERE)/out

.DEFAULT_GOAL := help

.PHONY: help
help: ## show this help
	@printf "misc/hf-glue — experimental hardfork testbed\n\n"
	@awk 'BEGIN {FS=":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: gen-local-genesis
gen-local-genesis: ## rebuild gnoland1 base genesis locally (runs misc/deployments/gnoland1/gen-genesis.sh); output at out/source-genesis.json
	@mkdir -p $(OUT)
	@OUT=$(OUT) REPO=$(REPO) $(HERE)/scripts/gen-local-genesis.sh

.PHONY: fetch
fetch: ## build hardfork tool + fetch genesis from $SOURCE into out/genesis.json
	@mkdir -p $(OUT)
	@SOURCE=$(SOURCE) \
	 ORIGINAL_CHAIN_ID=$(ORIGINAL_CHAIN_ID) \
	 CHAIN_ID=$(CHAIN_ID) \
	 HALT_HEIGHT=$(HALT_HEIGHT) \
	 OUT=$(OUT) REPO=$(REPO) \
	 $(HERE)/scripts/fetch.sh

.PHONY: init
init: ## generate single-validator secrets and patch out/genesis.json
	@mkdir -p $(OUT)
	@VALIDATOR_NAME=$(VALIDATOR_NAME) \
	 OUT=$(OUT) REPO=$(REPO) \
	 $(HERE)/scripts/init-node.sh

.PHONY: up
up: ## start the gnoland node in docker (requires fetch+init first)
	@test -f $(OUT)/genesis.json || { echo "missing out/genesis.json — run 'make fetch init' first"; exit 1; }
	@test -f $(OUT)/gnoland-home/secrets/priv_validator_key.json || { echo "missing secrets — run 'make init' first"; exit 1; }
	docker compose up -d --build
	@echo ""
	@echo "Node RPC:  http://localhost:26657"
	@echo "Chain ID:  $(CHAIN_ID)"
	@echo "Tail logs: make logs"

.PHONY: down
down: ## stop the node (keep state)
	docker compose down

.PHONY: logs
logs: ## tail node logs
	docker compose logs -f --tail=200

.PHONY: status
status: ## show node /status via RPC
	@curl -s http://localhost:26657/status | jq . || true

.PHONY: reset
reset: down ## stop and wipe ALL generated state (genesis, keys, db)
	rm -rf $(OUT)
	@echo "out/ removed."

.PHONY: smoketest
smoketest: ## run 'hardfork test' in-memory against out/genesis.json
	@test -f $(OUT)/genesis.json || { echo "missing out/genesis.json — run 'make fetch' first"; exit 1; }
	cd $(REPO)/misc/hardfork && go run . test --genesis $(OUT)/genesis.json --verbose
