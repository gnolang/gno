#!/usr/bin/env bash
# Fetch full source-chain state (genesis + historical txs) and build a
# hardforked genesis.json via the misc/hardfork tool.
#
# Two-stage pull (both remote):
#   1) Base genesis from $SOURCE (GitHub release asset by default, or an RPC
#      URL). For RPC URLs we unwrap the JSON-RPC envelope. Large genesis docs
#      often fail on RPC endpoints, hence the release-asset default.
#   2) Historical txs from $RPC_URL using contribs/tx-archive with batching.
#      Each tx gets block_height + chain_id metadata that the hardfork replay
#      needs.
#
# Both land in a staging dir laid out like a gnoland node data dir:
#   $OUT/source/
#     config/genesis.json   (from step 1)
#     txs.jsonl             (from step 2)
#
# The hardfork tool reads both via its dirSource.
#
# Inputs (env):
#   SOURCE              URL or local file of the base genesis.json
#                       (default: gnoland1 release asset)
#   RPC_URL             RPC endpoint to pull historical blocks from
#   ORIGINAL_CHAIN_ID   source chain ID
#   CHAIN_ID            new chain ID
#   HALT_HEIGHT         optional — block height to stop at. If empty, pulls
#                       up to the RPC's latest block at start time.
#   OUT                 output directory (absolute)
#   REPO                repo root (absolute)
set -euo pipefail

: "${SOURCE:?SOURCE is required}"
: "${RPC_URL:?RPC_URL is required}"
: "${ORIGINAL_CHAIN_ID:?ORIGINAL_CHAIN_ID is required}"
: "${CHAIN_ID:?CHAIN_ID is required}"
: "${OUT:?OUT is required}"
: "${REPO:?REPO is required}"

GENESIS="$OUT/genesis.json"
STAGE="$OUT/source"
STAGE_GEN="$STAGE/config/genesis.json"
STAGE_TXS="$STAGE/txs.jsonl"

banner() {
  printf '\n\033[1;36m━━━ %s ━━━\033[0m\n' "$1"
}

banner "fetch hardfork genesis"
printf "  base genesis:      \033[36m%s\033[0m\n" "$SOURCE"
printf "  rpc (for blocks):  \033[36m%s\033[0m\n" "$RPC_URL"
printf "  original chain id: \033[36m%s\033[0m\n" "$ORIGINAL_CHAIN_ID"
printf "  new chain id:      \033[36m%s\033[0m\n" "$CHAIN_ID"
printf "  halt height:       \033[36m%s\033[0m\n" "${HALT_HEIGHT:-<auto-detect>}"
printf "  output:            \033[36m%s\033[0m\n" "$GENESIS"
printf "  staging dir:       \033[36m%s\033[0m\n" "$STAGE"
echo ""

mkdir -p "$STAGE/config"

# ---------------------------------------------------------------------------
# Step 1: base genesis.json
# ---------------------------------------------------------------------------
banner "step 1/3 — base genesis"
if [[ -f "$STAGE_GEN" ]]; then
  echo "  already present ($(wc -c < "$STAGE_GEN" | tr -d ' ') bytes), skipping"
elif [[ "$SOURCE" == http://* || "$SOURCE" == https://* ]]; then
  if [[ "$SOURCE" == *.json || "$SOURCE" == */genesis.json ]]; then
    echo "  downloading from $SOURCE"
    curl -fSL --retry 3 --retry-delay 5 --max-time 600 --progress-bar \
      -o "$STAGE_GEN" "$SOURCE"
  else
    GENESIS_URL="${SOURCE%/}/genesis"
    echo "  downloading from $GENESIS_URL"
    curl -fSL --retry 3 --retry-delay 5 --max-time 600 --progress-bar \
      -o "$STAGE/envelope.json" "$GENESIS_URL"
    jq -c '.result.genesis' < "$STAGE/envelope.json" > "$STAGE_GEN"
    rm -f "$STAGE/envelope.json"
  fi
  echo "  $(wc -c < "$STAGE_GEN" | tr -d ' ') bytes"
elif [[ -f "$SOURCE" ]]; then
  echo "  copying local $SOURCE"
  cp "$SOURCE" "$STAGE_GEN"
else
  echo "ERROR: SOURCE is not a URL, not a file: $SOURCE" >&2
  exit 1
fi

# ---------------------------------------------------------------------------
# Step 2: historical txs via contribs/tx-archive
# ---------------------------------------------------------------------------
banner "step 2/3 — historical txs"
if [[ -z "${HALT_HEIGHT:-}" ]]; then
  HALT_HEIGHT=$(curl -fsS --max-time 30 "${RPC_URL%/}/status" \
    | jq -r '.result.sync_info.latest_block_height')
  echo "  halt height auto-detected: $HALT_HEIGHT"
fi

if [[ -f "$STAGE_TXS" ]]; then
  echo "  $STAGE_TXS already present ($(wc -l < "$STAGE_TXS" | tr -d ' ') txs), skipping"
else
  echo "  running tx-archive backup against $RPC_URL (1..$HALT_HEIGHT)"
  cd "$REPO/contribs/tx-archive"
  # No -verbose: rely on the new Info-level Progress lines that land every ~5s.
  go run ./cmd backup \
    -remote "$RPC_URL" \
    -from-block 1 \
    -to-block "$HALT_HEIGHT" \
    -batch 1000 \
    -output-path "$STAGE_TXS" \
    -overwrite
  echo "  $(wc -l < "$STAGE_TXS" | tr -d ' ') txs in $STAGE_TXS"
fi

# ---------------------------------------------------------------------------
# Step 3: assemble the hardfork genesis
# ---------------------------------------------------------------------------
banner "step 3/3 — assemble hardfork genesis"
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
