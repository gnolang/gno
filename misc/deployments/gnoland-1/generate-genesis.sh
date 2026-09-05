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
#   PV_KEY        path to the new validator's priv_validator_key.json.
#                 When set, a valset-reset migration tx is built and
#                 appended at the end of replay (updates r/sys/validators/v2
#                 to match the new GenesisDoc.Validators). Leave empty to
#                 skip migrations.
#   CALLER        govDAO T1 address that runs the migration MsgRun
#                 (default: g1manfred47...)
#
# Usage:
#   ./generate-genesis.sh
#   SOURCE=http://rpc.gno.land:26657 ./generate-genesis.sh
#   HALT_HEIGHT=704052 ./generate-genesis.sh
#   PV_KEY=./my-valkey.json ./generate-genesis.sh

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

# Build the post-replay migration jsonl if a new-valset priv_validator_key
# is provided. This appends a govDAO proposal tx (MsgRun) at the end of
# appState.Txs that resets r/sys/validators/v2 to match the new
# GenesisDoc.Validators — reconciling the in-gno side with the tm2 side.
if [[ -n "${PV_KEY:-}" ]]; then
    MIG_JSONL="$SCRIPT_DIR/migrations/migrations.jsonl"
    printf "Building migrations (PV_KEY=%s)...\n" "$PV_KEY"
    CALLER="${CALLER:-g1manfred47kzduec920z88wfr64ylksmdcedlf5}" \
    PV_KEY="$PV_KEY" \
    OUT_JSONL="$MIG_JSONL" \
    CHAIN_ID="$CHAIN_ID" \
    REPO_ROOT="$REPO_ROOT" \
      "$SCRIPT_DIR/migrations/build.sh"
    CMD_ARGS+=(--migration-tx "$MIG_JSONL")
fi

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
