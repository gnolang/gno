# misc/hf-glue — HIGHLY EXPERIMENTAL hardfork testbed
#
# See README.md. Do not merge.

SHELL := bash

# ---- configurable ----------------------------------------------------------
SOURCE             ?= https://github.com/gnolang/gno/releases/download/chain/gnoland1.0/genesis.json
RPC_URL            ?= https://rpc.gno.land
ORIGINAL_CHAIN_ID  ?= gnoland1
CHAIN_ID           ?= gnoland-1
HALT_HEIGHT        ?=
VALIDATOR_NAME     ?= hf-glue-local
# Manual validator identity (used by `make genesis`, not by `make init/up`).
# Pass bech32 strings instead of providing a priv_validator_key.json.
VALIDATOR_ADDR     ?=
VALIDATOR_PUBKEY   ?=
# Only used by `make fetch-from-dir` (alternative local-source flow).
NODE_DIR           ?=
TXS_JSONL          ?=
# Extra realm patches applied at hardfork-assembly time (see scripts/migrate.sh).
# Space-separated PKGPATH=SRCDIR entries. Set to empty to disable.
# (migrate.sh always patches r/sys/params from examples/; use this for more.)
PATCH_REALMS       ?=
# Optional: rotate the govDAO sole T1 member to a different address as part
# of the post-replay migration (see scripts/migrate.sh step 5 and
# misc/deployments/gnoland-1/migrations/build.sh). Empty = no rotation.
NEW_T1_ADDR        ?=
T1_PORTFOLIO       ?=
T1_WITHDRAW_REASON ?=

export SOURCE RPC_URL ORIGINAL_CHAIN_ID CHAIN_ID HALT_HEIGHT VALIDATOR_NAME VALIDATOR_ADDR VALIDATOR_PUBKEY NODE_DIR TXS_JSONL PATCH_REALMS NEW_T1_ADDR T1_PORTFOLIO T1_WITHDRAW_REASON

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

.PHONY: fetch migrate
fetch: migrate ## alias for `make migrate` (kept for back-compat)

migrate: ## run scripts/migrate.sh — declarative hardfork recipe (fetch + patch + assemble)
	@mkdir -p $(OUT)
	@SOURCE=$(SOURCE) \
	 RPC_URL=$(RPC_URL) \
	 ORIGINAL_CHAIN_ID=$(ORIGINAL_CHAIN_ID) \
	 CHAIN_ID=$(CHAIN_ID) \
	 HALT_HEIGHT=$(HALT_HEIGHT) \
	 PATCH_REALMS='$(PATCH_REALMS)' \
	 OUT=$(OUT) REPO=$(REPO) \
	 $(HERE)/scripts/migrate.sh

.PHONY: genesis
genesis: ## produce out/genesis.json only, from manual VALIDATOR_ADDR+VALIDATOR_PUBKEY (no secrets, no docker)
	@test -n "$(VALIDATOR_ADDR)"   || { echo "VALIDATOR_ADDR required (bech32 g1...)"; exit 1; }
	@test -n "$(VALIDATOR_PUBKEY)" || { echo "VALIDATOR_PUBKEY required (bech32 gpub1...)"; exit 1; }
	@test -n "$(HALT_HEIGHT)"      || { echo "HALT_HEIGHT required"; exit 1; }
	@mkdir -p $(OUT)
	@SOURCE=$(SOURCE) \
	 RPC_URL=$(RPC_URL) \
	 ORIGINAL_CHAIN_ID=$(ORIGINAL_CHAIN_ID) \
	 CHAIN_ID=$(CHAIN_ID) \
	 HALT_HEIGHT=$(HALT_HEIGHT) \
	 VALIDATOR_ADDR=$(VALIDATOR_ADDR) \
	 VALIDATOR_PUBKEY=$(VALIDATOR_PUBKEY) \
	 VALIDATOR_NAME=$(VALIDATOR_NAME) \
	 PATCH_REALMS='$(PATCH_REALMS)' \
	 OUT=$(OUT) REPO=$(REPO) \
	 $(HERE)/scripts/migrate.sh
	@go run -C $(REPO)/misc/hf-glue/fixvalidator . \
	 --address $(VALIDATOR_ADDR) --pubkey $(VALIDATOR_PUBKEY) \
	 --genesis $(OUT)/genesis.json \
	 --name $(VALIDATOR_NAME) --power 10
	@echo ""
	@echo "Genesis ready: $(OUT)/genesis.json"
	@echo "  SHA256:     $$(shasum -a 256 $(OUT)/genesis.json | cut -d' ' -f1)"
	@echo "  tx count:   $$(jq '.app_state.txs | length' $(OUT)/genesis.json)"
	@echo "  chain_id:   $$(jq -r '.chain_id' $(OUT)/genesis.json)"

.PHONY: fetch-from-dir
fetch-from-dir: ## (alt) build hardfork genesis from a local gnoland data dir ($NODE_DIR, $TXS_JSONL, $HALT_HEIGHT required)
	@mkdir -p $(OUT)
	@NODE_DIR=$(NODE_DIR) TXS_JSONL=$(TXS_JSONL) \
	 ORIGINAL_CHAIN_ID=$(ORIGINAL_CHAIN_ID) \
	 CHAIN_ID=$(CHAIN_ID) \
	 HALT_HEIGHT=$(HALT_HEIGHT) \
	 OUT=$(OUT) REPO=$(REPO) \
	 $(HERE)/scripts/fetch-from-dir.sh

.PHONY: init
init: ## generate single-validator secrets and patch out/genesis.json
	@mkdir -p $(OUT)
	@VALIDATOR_NAME=$(VALIDATOR_NAME) \
	 OUT=$(OUT) REPO=$(REPO) \
	 $(HERE)/scripts/init-node.sh

.PHONY: up
up: init ## start the gnoland node in docker (init is idempotent and re-stages genesis if changed)
	docker compose up -d --build
	@echo ""
	@echo "Node RPC:  http://localhost:26657"
	@echo "Chain ID:  $(CHAIN_ID)"
	@echo "Tail logs: make logs"

.PHONY: down
down: ## stop the node (keep state)
	docker compose down

.PHONY: reset-db
reset-db: down ## wipe node db + WAL + last-signed state, keep genesis + keys
	rm -rf $(OUT)/gnoland-home/db $(OUT)/gnoland-home/wal $(OUT)/gnoland-home/data
	printf '{"height":"0","round":"0","step":0}\n' > $(OUT)/gnoland-home/secrets/priv_validator_state.json
	@echo "db wiped; next 'make up' will re-replay genesis from scratch"

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
smoketest: ## run 'gnogenesis fork test' in-memory against out/genesis.json
	@test -f $(OUT)/genesis.json || { echo "missing out/genesis.json — run 'make fetch' first"; exit 1; }
	cd $(REPO)/contribs/gnogenesis && go run . fork test --genesis $(OUT)/genesis.json --verbose

.PHONY: replay-log
replay-log: ## run in-process genesis replay, tee full log to out/replay.log
	@$(HERE)/scripts/replay-log.sh

.PHONY: report-replay
report-replay: ## generate out/REPLAY-REPORT.md from out/replay.log
	@$(HERE)/scripts/report-replay.sh

.PHONY: check-state
check-state: ## probe running node, compare to gno.land, write out/STATE-REPORT.md
	@$(HERE)/scripts/check-state.sh

.PHONY: reports
reports: replay-log report-replay check-state ## full reporting pipeline (replay + compare + write)
	@echo ""
	@echo "Reports:"
	@echo "  $(OUT)/replay.log"
	@echo "  $(OUT)/REPLAY-REPORT.md"
	@echo "  $(OUT)/STATE-REPORT.md"
