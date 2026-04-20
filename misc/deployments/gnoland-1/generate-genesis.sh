#!/usr/bin/env bash
# Generate gnoland-1 hardfork genesis.
#
# Wraps `gnogenesis fork generate` with gnoland-1 chain IDs hardcoded.
#
# Env vars:
#   SOURCE        source to fetch state from (default: production RPC)
#                 http://...       RPC of a running or recently-halted node
#                 /path/to/dir     local node data directory (stopped node)
#                 /path/to/*.json  exported genesis
#   HALT_HEIGHT   block height at which gnoland1 was halted
#                 (empty = auto-detect from source)
#
# Usage:
#   ./generate-genesis.sh
#   SOURCE=http://rpc.gno.land:26657 ./generate-genesis.sh
#   HALT_HEIGHT=704052 ./generate-genesis.sh

set -euo pipefail

CHAIN_ID="gnoland-1"
ORIGINAL_CHAIN_ID="gnoland1"

SOURCE="${SOURCE:-http://rpc.gno.land:26657}"
HALT_HEIGHT="${HALT_HEIGHT:-}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
GNOGENESIS_DIR="$REPO_ROOT/contribs/gnogenesis"
OUTPUT="$SCRIPT_DIR/genesis.json"

# Build the gnogenesis binary if not already available.
if command -v gnogenesis >/dev/null 2>&1; then
    BIN="gnogenesis"
else
    BIN="$SCRIPT_DIR/genesis-work/bin/gnogenesis"
    if [[ ! -x "$BIN" ]]; then
        printf "Building gnogenesis...\n"
        mkdir -p "$(dirname "$BIN")"
        go build -C "$GNOGENESIS_DIR" -o "$BIN" .
    fi
fi

CMD_ARGS=(
    fork generate
    --source "$SOURCE"
    --chain-id "$CHAIN_ID"
    --original-chain-id "$ORIGINAL_CHAIN_ID"
    --output "$OUTPUT"
)
[[ -n "$HALT_HEIGHT" ]] && CMD_ARGS+=(--halt-height "$HALT_HEIGHT")

"$BIN" "${CMD_ARGS[@]}"

# Print sha256 for cross-validator coordination.
if [[ -f "$OUTPUT" ]]; then
    if command -v sha256sum >/dev/null 2>&1; then
        SHA256=$(sha256sum "$OUTPUT" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        SHA256=$(shasum -a 256 "$OUTPUT" | cut -d' ' -f1)
    fi
    printf "\nsha256: %s\n" "${SHA256:-<sha256 tool not found>}"
fi
