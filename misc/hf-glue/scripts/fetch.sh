#!/usr/bin/env bash
# Fetch source-chain state and build a hardforked genesis.json via the
# misc/hardfork tool (shipped by PR #5511).
#
# For large chains (betanet/gnoland1), the genesis doc can be 100+ MB and the
# JSON-RPC /genesis endpoint often fails under load. This script detects RPC
# sources, pre-downloads the base genesis via curl, and feeds the local file to
# the hardfork tool with --skip-txs.  Historical txs are fetched separately via
# RPC block-by-block (which works fine at any chain size).
#
# Inputs (env):
#   SOURCE              RPC URL / local data dir / exported file
#   ORIGINAL_CHAIN_ID   source chain ID (for historical signature verification)
#   CHAIN_ID            new chain ID
#   HALT_HEIGHT         optional — block height to stop at
#   OUT                 output directory (absolute)
#   REPO                repo root (absolute)
set -euo pipefail

: "${SOURCE:?SOURCE is required}"
: "${ORIGINAL_CHAIN_ID:?ORIGINAL_CHAIN_ID is required}"
: "${CHAIN_ID:?CHAIN_ID is required}"
: "${OUT:?OUT is required}"
: "${REPO:?REPO is required}"

GENESIS="$OUT/genesis.json"

echo "── fetch hardfork genesis ───────────────────────────────────"
echo "  source:            $SOURCE"
echo "  original chain id: $ORIGINAL_CHAIN_ID"
echo "  new chain id:      $CHAIN_ID"
echo "  halt height:       ${HALT_HEIGHT:-<auto-detect>}"
echo "  output:            $GENESIS"
echo ""

# ---------------------------------------------------------------------------
# For HTTP(S) sources, pre-download the genesis file locally.
# Supports two URL forms:
#   1. Direct genesis.json URL (e.g. GitHub release asset) — used as-is
#   2. RPC endpoint URL — appends /genesis, extracts from JSON-RPC envelope
# ---------------------------------------------------------------------------
EFFECTIVE_SOURCE="$SOURCE"

if [[ "$SOURCE" == http://* || "$SOURCE" == https://* ]]; then
  BASE_GENESIS="$OUT/source-genesis.json"

  if [[ -f "$BASE_GENESIS" ]]; then
    echo "  reusing cached source genesis: $BASE_GENESIS"
  else
    # Detect whether this is a direct genesis.json URL or an RPC endpoint.
    if [[ "$SOURCE" == *.json || "$SOURCE" == */genesis.json ]]; then
      # Direct download (e.g. GitHub release asset).
      echo "  downloading genesis from $SOURCE ..."
      echo "  (this may take a few minutes for large files)"
      curl -fSL --retry 3 --retry-delay 5 --max-time 600 \
        -o "$BASE_GENESIS" \
        "$SOURCE"
    else
      # RPC endpoint — fetch via /genesis and unwrap JSON-RPC envelope.
      GENESIS_URL="${SOURCE%/}/genesis"
      echo "  downloading base genesis from $GENESIS_URL ..."
      echo "  (this may take a few minutes for large chains)"
      curl -fSL --retry 3 --retry-delay 5 --max-time 600 \
        -o "$OUT/source-genesis-envelope.json" \
        "$GENESIS_URL"
      echo "  extracting genesis from JSON-RPC envelope..."
      jq -c '.result.genesis' < "$OUT/source-genesis-envelope.json" > "$BASE_GENESIS"
      rm -f "$OUT/source-genesis-envelope.json"
    fi

    SIZE=$(wc -c < "$BASE_GENESIS" | tr -d ' ')
    echo "  source genesis: $(echo "scale=1; $SIZE / 1048576" | bc) MB"
  fi

  EFFECTIVE_SOURCE="$BASE_GENESIS"
fi

cd "$REPO/misc/hardfork"

ARGS=(
  genesis
  --source "$EFFECTIVE_SOURCE"
  --chain-id "$CHAIN_ID"
  --original-chain-id "$ORIGINAL_CHAIN_ID"
  --output "$GENESIS"
)

if [[ -n "${HALT_HEIGHT:-}" ]]; then
  ARGS+=(--halt-height "$HALT_HEIGHT")
fi

# When using a pre-downloaded genesis file as source, historical txs are
# already embedded in the genesis app_state.txs — no need to fetch separately.
go run . "${ARGS[@]}"

echo ""
if command -v sha256sum >/dev/null 2>&1; then
  echo "sha256: $(sha256sum "$GENESIS" | cut -d' ' -f1)"
elif command -v shasum >/dev/null 2>&1; then
  echo "sha256: $(shasum -a 256 "$GENESIS" | cut -d' ' -f1)"
fi
echo "done — genesis written to $GENESIS"
