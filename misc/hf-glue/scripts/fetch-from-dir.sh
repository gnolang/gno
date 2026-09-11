#!/usr/bin/env bash
# Alternative to fetch.sh that pulls source-chain state from a LOCAL gnoland
# node data directory instead of hitting RPC endpoints.
#
# Use this when you have a locally-synced gnoland1 node — replaying its
# blockstore.db is far faster (and offline) than pulling blocks via RPC.
#
# Expected layout of $NODE_DIR (matches `gnoland start --data-dir` output):
#   $NODE_DIR/
#     config/genesis.json
#     db/
#       blockstore.db/      ← historical txs live here
#       state.db/
#       ...
#
# Currently the misc/hardfork tool's dirSource reads genesis.json + an optional
# txs.jsonl. Reading blockstore.db directly is not yet implemented, so this
# script first runs tx-archive against a locally-spawned RPC (or — future —
# reads the block store directly once misc/hardfork/source_dir.go grows that
# support). For now we assume the caller also provides a txs.jsonl alongside
# the genesis, or has run an RPC locally.
#
# Inputs (env):
#   NODE_DIR            path to a gnoland-data dir
#   TXS_JSONL           optional — path to a pre-exported txs.jsonl
#                       (if absent, an error is raised — see TODO below)
#   ORIGINAL_CHAIN_ID   source chain ID
#   CHAIN_ID            new chain ID
#   HALT_HEIGHT         required for local mode (we can't auto-detect without RPC)
#   OUT                 output directory (absolute)
#   REPO                repo root (absolute)
set -euo pipefail

: "${NODE_DIR:?NODE_DIR is required (path to a gnoland data directory)}"
: "${ORIGINAL_CHAIN_ID:?ORIGINAL_CHAIN_ID is required}"
: "${CHAIN_ID:?CHAIN_ID is required}"
: "${HALT_HEIGHT:?HALT_HEIGHT is required when using local source}"
: "${OUT:?OUT is required}"
: "${REPO:?REPO is required}"

GENESIS="$OUT/genesis.json"
STAGE="$OUT/source"
STAGE_GEN="$STAGE/config/genesis.json"
STAGE_TXS="$STAGE/txs.jsonl"

echo "── fetch hardfork genesis (local dir) ────────────────────────"
echo "  node dir:          $NODE_DIR"
echo "  original chain id: $ORIGINAL_CHAIN_ID"
echo "  new chain id:      $CHAIN_ID"
echo "  halt height:       $HALT_HEIGHT"
echo "  output:            $GENESIS"
echo ""

mkdir -p "$STAGE/config"

# ---- genesis.json from the local node ----
SRC_GEN="$NODE_DIR/config/genesis.json"
if [[ ! -f "$SRC_GEN" ]]; then
  SRC_GEN="$NODE_DIR/genesis.json"
fi
if [[ ! -f "$SRC_GEN" ]]; then
  echo "ERROR: genesis.json not found under $NODE_DIR" >&2
  exit 1
fi
echo "[1/3] using base genesis: $SRC_GEN"
cp "$SRC_GEN" "$STAGE_GEN"

# ---- txs.jsonl ----
if [[ -n "${TXS_JSONL:-}" ]]; then
  if [[ ! -f "$TXS_JSONL" ]]; then
    echo "ERROR: TXS_JSONL=$TXS_JSONL does not exist" >&2
    exit 1
  fi
  echo "[2/3] using provided txs.jsonl: $TXS_JSONL"
  cp "$TXS_JSONL" "$STAGE_TXS"
else
  # TODO: once misc/hardfork/source_dir.go reads blockstore.db directly,
  # point dirSource at $NODE_DIR/db/blockstore.db and skip this step.
  echo "ERROR: no TXS_JSONL provided." >&2
  echo "  Either (a) pass TXS_JSONL=/path/to/txs.jsonl, or" >&2
  echo "  (b) run 'contribs/tx-archive backup' against a local RPC on this node," >&2
  echo "  (c) wait for misc/hardfork to grow blockstore.db support (open issue)." >&2
  exit 1
fi

# ---- assemble ----
echo ""
echo "[3/3] assembling hardfork genesis..."
cd "$REPO/misc/hardfork"

ARGS=(
  genesis
  --source "$STAGE"
  --chain-id "$CHAIN_ID"
  --original-chain-id "$ORIGINAL_CHAIN_ID"
  --halt-height "$HALT_HEIGHT"
  --output "$GENESIS"
)
go run . "${ARGS[@]}"

echo ""
if command -v sha256sum >/dev/null 2>&1; then
  echo "sha256: $(sha256sum "$GENESIS" | cut -d' ' -f1)"
elif command -v shasum >/dev/null 2>&1; then
  echo "sha256: $(shasum -a 256 "$GENESIS" | cut -d' ' -f1)"
fi
echo "done — genesis written to $GENESIS"
