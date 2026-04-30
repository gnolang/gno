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
# Optional: keyless-validator mode — bech32 address + pubkey of the
# remote-signer key (e.g. gnokms). Set both to skip priv_validator_key.json
# generation and bake the supplied identity into both the GenesisDoc valset
# and the post-fork r/sys/validators/v2 reset migration.
VALIDATOR_ADDRESS  ?=
VALIDATOR_PUBKEY   ?=

export SOURCE RPC_URL ORIGINAL_CHAIN_ID CHAIN_ID HALT_HEIGHT VALIDATOR_NAME NODE_DIR TXS_JSONL PATCH_REALMS NEW_T1_ADDR T1_PORTFOLIO T1_WITHDRAW_REASON VALIDATOR_ADDRESS VALIDATOR_PUBKEY

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

.PHONY: fetch genesis migrate
fetch: migrate ## alias for `make migrate` (kept for back-compat)
genesis: migrate ## alias for `make migrate` — name used by verify-reproducibility.sh

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
	cd $(REPO)/contribs/gnogenesis && go run . fork test --genesis $(OUT)/genesis.json --skip-failing-genesis-txs --verbose

# ---- cluster (2-node docker) ---------------------------------------------
NODES ?= 2
CLUSTER_COMPOSE := -f docker-compose.cluster.yml

.PHONY: cluster-init
cluster-init: ## generate N=$(NODES) validator homes under out/cluster/ and rewrite genesis
	@test -f $(OUT)/genesis.json || { echo "missing out/genesis.json — run 'make fetch' first"; exit 1; }
	@NODES=$(NODES) VALIDATOR_NAME=$(VALIDATOR_NAME) \
	 OUT=$(OUT) REPO=$(REPO) \
	 $(HERE)/scripts/init-cluster.sh

.PHONY: cluster-up
cluster-up: ## start the N-node cluster in docker (requires fetch + cluster-init)
	@test -d $(OUT)/cluster/node0 || { echo "missing out/cluster/ — run 'make cluster-init' first"; exit 1; }
	docker compose $(CLUSTER_COMPOSE) up -d --build
	@echo ""
	@echo "node0 RPC:  http://localhost:26657"
	@echo "node1 RPC:  http://localhost:36657"
	@echo "gnoweb:     http://localhost:8888"
	@echo "Tail logs:  make cluster-logs"

.PHONY: cluster-down
cluster-down: ## stop the cluster (keep state)
	docker compose $(CLUSTER_COMPOSE) down

.PHONY: cluster-logs
cluster-logs: ## tail cluster logs
	docker compose $(CLUSTER_COMPOSE) logs -f --tail=200

.PHONY: cluster-status
cluster-status: ## check both nodes' /status and peer connectivity
	@printf "── node0 /status ─────────────────────\n"
	@curl -s http://localhost:26657/status | jq '.result.sync_info' || true
	@printf "\n── node1 /status ─────────────────────\n"
	@curl -s http://localhost:36657/status | jq '.result.sync_info' || true
	@printf "\n── node0 /net_info (peers=%s) ────────\n" "$$(curl -s http://localhost:26657/net_info | jq -r '.result.n_peers')"
	@curl -s http://localhost:26657/net_info | jq '.result.peers[].node_info | {id: .id, moniker: .moniker}' || true
	@printf "\n── node1 /net_info (peers=%s) ────────\n" "$$(curl -s http://localhost:36657/net_info | jq -r '.result.n_peers')"
	@curl -s http://localhost:36657/net_info | jq '.result.peers[].node_info | {id: .id, moniker: .moniker}' || true

.PHONY: cluster-reset-db
cluster-reset-db: cluster-down ## wipe cluster db + WAL, keep genesis + keys
	@for d in $(OUT)/cluster/node*; do \
	  rm -rf "$$d/db" "$$d/wal" "$$d/data"; \
	  printf '{"height":"0","round":"0","step":0}\n' > "$$d/secrets/priv_validator_state.json"; \
	done
	@echo "cluster db wiped; next 'make cluster-up' will re-replay from scratch"

.PHONY: cluster-reset
cluster-reset: cluster-down ## stop cluster and wipe cluster state (keep out/genesis.json)
	rm -rf $(OUT)/cluster
	@echo "out/cluster/ removed."

.PHONY: assert-migrations
assert-migrations: ## assert live node's post-replay state matches migration intent (overrides: REMOTE= EXPECTED_T1= EXPECTED_VALSET_ADDRS=)
	@REMOTE="$${REMOTE:-http://localhost:26657}" \
	 EXPECTED_T1="$${EXPECTED_T1:-$(NEW_T1_ADDR)}" \
	 EXPECTED_VALSET_ADDRS="$${EXPECTED_VALSET_ADDRS:-$(VALIDATOR_ADDR)}" \
	 $(HERE)/scripts/assert-migrations.sh

.PHONY: verify-txs-jsonl
verify-txs-jsonl: ## compare out/source/txs.jsonl against source-chain RPC (cardinality + spot-check at random heights with txs)
	@test -n "$(HALT_HEIGHT)" || { echo "HALT_HEIGHT required"; exit 1; }
	@RPC_URL="$(RPC_URL)" \
	 HALT_HEIGHT=$(HALT_HEIGHT) \
	 OUT=$(OUT) \
	 $(HERE)/scripts/verify-txs-jsonl.sh

.PHONY: state-diff
state-diff: ## diff realm render output source-chain@halt_height vs replay (overrides: SOURCE_RPC= REPLAY_RPC= REALMS= REPLAY_HEIGHT=); writes out/STATE-DIFF.md
	@test -n "$(HALT_HEIGHT)" || { echo "HALT_HEIGHT required"; exit 1; }
	@SOURCE_RPC="$${SOURCE_RPC:-$(RPC_URL)}" \
	 REPLAY_RPC="$${REPLAY_RPC:-http://localhost:26657}" \
	 HALT_HEIGHT=$(HALT_HEIGHT) \
	 REALMS="$${REALMS:-}" \
	 OUT=$(OUT) \
	 $(HERE)/scripts/state-diff.sh

.PHONY: audit-balances
audit-balances: ## diff per-signer ugnot balance source-chain@halt_height vs replay (overrides: SOURCE_RPC= REPLAY_RPC= SIGNER_LIMIT=); writes out/BALANCE-AUDIT.md
	@test -n "$(HALT_HEIGHT)" || { echo "HALT_HEIGHT required"; exit 1; }
	@SOURCE_RPC="$${SOURCE_RPC:-$(RPC_URL)}" \
	 REPLAY_RPC="$${REPLAY_RPC:-http://localhost:26657}" \
	 HALT_HEIGHT=$(HALT_HEIGHT) \
	 SIGNER_LIMIT="$${SIGNER_LIMIT:-200}" \
	 OUT=$(OUT) \
	 $(HERE)/scripts/audit-balances.sh

.PHONY: verify-reproducibility
verify-reproducibility: ## build genesis twice in isolated OUT dirs and assert identical SHA256 (forwards all `make genesis` envs)
	@test -n "$(HALT_HEIGHT)" || { echo "HALT_HEIGHT required"; exit 1; }
	@SHARED_TXS="$${SHARED_TXS:-$(OUT)/source/txs.jsonl}" \
	 OUT=$(OUT) \
	 $(HERE)/scripts/verify-reproducibility.sh

.PHONY: audit-realm-imports
audit-realm-imports: ## flag addpkgs whose import paths no longer resolve against the current stdlib+examples tree; writes out/REALM-IMPORTS-AUDIT.md
	@REPO=$(REPO) OUT=$(OUT) \
	 $(HERE)/scripts/audit-realm-imports.sh

.PHONY: compare-gas-modes
compare-gas-modes: ## run smoketest twice with gas_replay_mode=strict vs source and diff failure counts; writes out/GAS-MODES-COMPARE.md
	@test -f $(OUT)/genesis.json || { echo "missing out/genesis.json — run 'make genesis' first"; exit 1; }
	@REPO=$(REPO) OUT=$(OUT) \
	 $(HERE)/scripts/compare-gas-modes.sh

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
