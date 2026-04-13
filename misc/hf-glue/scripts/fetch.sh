#!/usr/bin/env bash
# Fetch source-chain state and build a hardforked genesis.json via the
# misc/hardfork tool (shipped by PR #5411).
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

cd "$REPO/misc/hardfork"

ARGS=(
  genesis
  --source "$SOURCE"
  --chain-id "$CHAIN_ID"
  --original-chain-id "$ORIGINAL_CHAIN_ID"
  --output "$GENESIS"
)

if [[ -n "${HALT_HEIGHT:-}" ]]; then
  ARGS+=(--halt-height "$HALT_HEIGHT")
fi

go run . "${ARGS[@]}"

echo ""
if command -v sha256sum >/dev/null 2>&1; then
  echo "sha256: $(sha256sum "$GENESIS" | cut -d' ' -f1)"
elif command -v shasum >/dev/null 2>&1; then
  echo "sha256: $(shasum -a 256 "$GENESIS" | cut -d' ' -f1)"
fi
echo "done — genesis written to $GENESIS"
