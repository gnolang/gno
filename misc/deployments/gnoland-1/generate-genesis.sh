#!/usr/bin/env bash
# Generate gnoland-1 genesis from an existing gnoland1 node.
#
# Three phases:
#   Phase 1: Extract the base genesis from the local gnoland1 state
#   Phase 2: Apply the upgrade overlay (new contracts, config changes)
#   Phase 3: Export and append historical transactions with metadata
#
# Prerequisites:
#   - gnoland, gnogenesis, gnokey binaries (built from this branch)
#   - tx-archive binary (github.com/gnolang/tx-archive)
#   - A stopped gnoland1 node with its data directory
#
# Usage:
#   ./generate-genesis.sh --data-dir /path/to/gnoland1-data [OPTIONS]
#
# Options:
#   --data-dir PATH       gnoland1 node data directory (required)
#   --rpc URL             gnoland1 RPC for tx export (default: http://127.0.0.1:26657)
#   --halt-height N       block height where gnoland1 was halted
#   --output PATH         output genesis file (default: genesis.json)
#   --overlay-dir PATH    directory with overlay scripts (default: ./overlay/)
#   --skip-phase N        skip phase N (1, 2, or 3)
#   --debug               print every command

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

CHAIN_ID="gnoland-1"
ORIGINAL_CHAIN_ID="gnoland1"

# Defaults
DATA_DIR=""
RPC_URL="${RPC_URL:-http://127.0.0.1:26657}"
HALT_HEIGHT="${HALT_HEIGHT:-}"
OUTPUT="${OUTPUT:-genesis.json}"
OVERLAY_DIR=""
SKIP_PHASES=()
DEBUG=false

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
OVERLAY_DIR="${OVERLAY_DIR:-$SCRIPT_DIR/overlay}"
WORK_DIR="$SCRIPT_DIR/genesis-work"

# =============================================================================
# Parse flags
# =============================================================================

while [[ $# -gt 0 ]]; do
    case $1 in
        --data-dir)    DATA_DIR="$2"; shift 2 ;;
        --rpc)         RPC_URL="$2"; shift 2 ;;
        --halt-height) HALT_HEIGHT="$2"; shift 2 ;;
        --output)      OUTPUT="$2"; shift 2 ;;
        --overlay-dir) OVERLAY_DIR="$2"; shift 2 ;;
        --skip-phase)  SKIP_PHASES+=("$2"); shift 2 ;;
        --debug)       DEBUG=true; shift ;;
        -h|--help)
            sed -n '2,/^$/s/^# //p' "$0"
            exit 0
            ;;
        *) echo "Unknown option: $1" >&2; exit 1 ;;
    esac
done

run() {
    if [ "$DEBUG" = true ]; then
        printf "    \033[2m\$ %s\033[0m\n" "$*" >&2
    fi
    "$@"
}

should_skip() { printf '%s\n' "${SKIP_PHASES[@]}" 2>/dev/null | grep -qx "$1"; }

sha256_file() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | cut -d' ' -f1
    elif command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$1" | cut -d' ' -f1
    else
        echo "ERROR: no sha256 tool found" >&2; return 1
    fi
}

# =============================================================================
# Validation
# =============================================================================

if [ -z "$DATA_DIR" ]; then
    echo "ERROR: --data-dir is required (path to gnoland1 node data)" >&2
    echo "Run with -h for usage." >&2
    exit 1
fi

if [ ! -d "$DATA_DIR" ]; then
    echo "ERROR: data directory does not exist: $DATA_DIR" >&2
    exit 1
fi

mkdir -p "$WORK_DIR"

echo "=== Configuration ==="
echo "  Data dir:          $DATA_DIR"
echo "  RPC URL:           $RPC_URL"
echo "  Halt height:       ${HALT_HEIGHT:-<auto-detect>}"
echo "  Output:            $OUTPUT"
echo "  Chain ID:          $CHAIN_ID"
echo "  Original Chain ID: $ORIGINAL_CHAIN_ID"
echo "  Overlay dir:       $OVERLAY_DIR"
echo ""

# =============================================================================
# Phase 1: Extract base genesis from local gnoland1 state
# =============================================================================
# Takes the existing genesis.json from the stopped node as the base.
# This preserves: balances, auth state, bank state, VM state, validators,
# and the original genesis txs (package deploys, setup scripts).
# =============================================================================

if ! should_skip 1; then
    printf "=== Phase 1: Extract base genesis from local state ===\n"

    # Find the genesis file from the stopped node.
    ORIGINAL_GENESIS=""
    for candidate in \
        "$DATA_DIR/genesis.json" \
        "$DATA_DIR/config/genesis.json" \
    ; do
        if [ -f "$candidate" ]; then
            ORIGINAL_GENESIS="$candidate"
            break
        fi
    done

    if [ -z "$ORIGINAL_GENESIS" ]; then
        echo "ERROR: Could not find genesis.json in $DATA_DIR" >&2
        exit 1
    fi

    printf "  Found genesis: %s\n" "$ORIGINAL_GENESIS"

    # Copy as our working base.
    BASE_GENESIS="$WORK_DIR/base-genesis.json"
    cp "$ORIGINAL_GENESIS" "$BASE_GENESIS"

    # Extract halt height from the genesis if not specified.
    if [ -z "$HALT_HEIGHT" ]; then
        # Try to get it from the node's state (last committed block).
        # Fall back to user needing to specify it.
        printf "  Auto-detecting halt height from node state...\n"
        # TODO: read from blockstore.db or state.db
        echo "  WARNING: Could not auto-detect halt height. Specify --halt-height." >&2
    fi

    printf "  Base genesis extracted (%s)\n" "$(du -h "$BASE_GENESIS" | cut -f1)"
    printf "\n"
else
    printf "=== Phase 1: SKIPPED ===\n\n"
    BASE_GENESIS="$WORK_DIR/base-genesis.json"
fi

# =============================================================================
# Phase 2: Apply upgrade overlay
# =============================================================================
# Modifies the base genesis for the new chain:
#   - Update chain_id to gnoland-1
#   - Set original_chain_id for signature verification during tx replay
#   - Set initial_height to halt_height + 1
#   - Deploy new/updated contracts (overlay packages)
#   - Apply parameter changes (e.g., valoper min_fee)
#   - Run migration scripts (e.g., CLA setup)
#
# Overlay scripts are executed in alphabetical order from $OVERLAY_DIR.
# Each script receives the working genesis path as $1 and can modify it.
# =============================================================================

if ! should_skip 2; then
    printf "=== Phase 2: Apply upgrade overlay ===\n"

    OVERLAY_GENESIS="$WORK_DIR/overlay-genesis.json"
    cp "$BASE_GENESIS" "$OVERLAY_GENESIS"

    # Update chain ID and set upgrade fields.
    printf "  Setting chain_id=%s, original_chain_id=%s\n" "$CHAIN_ID" "$ORIGINAL_CHAIN_ID"
    # Use python3 for JSON manipulation (jq may not be available everywhere).
    python3 -c "
import json, sys
with open('$OVERLAY_GENESIS') as f:
    doc = json.load(f)
doc['chain_id'] = '$CHAIN_ID'
if 'app_state' not in doc:
    doc['app_state'] = {}
doc['app_state']['original_chain_id'] = '$ORIGINAL_CHAIN_ID'
if '$HALT_HEIGHT':
    height = int('${HALT_HEIGHT:-0}')
    if height > 0:
        doc['initial_height'] = height + 1
        doc['app_state']['initial_height'] = height + 1
with open('$OVERLAY_GENESIS', 'w') as f:
    json.dump(doc, f, indent=2)
"

    if [ -n "$HALT_HEIGHT" ] && [ "$HALT_HEIGHT" -gt 0 ]; then
        printf "  Set initial_height=%d\n" "$((HALT_HEIGHT + 1))"
    fi

    # Run overlay scripts if directory exists.
    if [ -d "$OVERLAY_DIR" ]; then
        overlay_scripts=("$OVERLAY_DIR"/*.sh)
        if [ -e "${overlay_scripts[0]}" ]; then
            printf "  Running %d overlay scripts:\n" "${#overlay_scripts[@]}"
            for script in "${overlay_scripts[@]}"; do
                printf "    -> %s\n" "$(basename "$script")"
                run bash "$script" "$OVERLAY_GENESIS"
            done
        else
            printf "  No overlay scripts found in %s\n" "$OVERLAY_DIR"
        fi
    else
        printf "  No overlay directory at %s (skipping)\n" "$OVERLAY_DIR"
    fi

    printf "  Overlay applied (%s)\n" "$(du -h "$OVERLAY_GENESIS" | cut -f1)"
    printf "\n"
else
    printf "=== Phase 2: SKIPPED ===\n\n"
    OVERLAY_GENESIS="$WORK_DIR/overlay-genesis.json"
fi

# =============================================================================
# Phase 3: Export and append historical transactions
# =============================================================================
# Exports all successful txs from gnoland1 (via tx-archive) with full metadata
# (block_height, timestamp, chain_id) and appends them to the genesis app_state.
# These txs will be replayed during InitChain with their original context.
# =============================================================================

if ! should_skip 3; then
    printf "=== Phase 3: Export and append historical transactions ===\n"

    TX_EXPORT="$WORK_DIR/historical-txs.jsonl"

    printf "  Exporting transactions from %s...\n" "$RPC_URL"
    tx-archive backup \
        --remote "$RPC_URL" \
        --output-path "$TX_EXPORT" \
        --from-block 1 \
        ${HALT_HEIGHT:+--to-block "$HALT_HEIGHT"} \
        --skip-failed-txs

    tx_count=$(wc -l < "$TX_EXPORT" | tr -d ' ')
    printf "  Exported %s transactions\n" "$tx_count"

    # Append historical txs to the genesis app_state.txs array.
    # The existing txs (from Phase 1 base genesis) are kept as-is (no metadata = genesis mode).
    # Historical txs have metadata with block_height > 0, so they replay in normal mode.
    printf "  Appending historical txs to genesis...\n"

    FINAL_GENESIS="$WORK_DIR/final-genesis.json"
    python3 -c "
import json, sys

with open('$OVERLAY_GENESIS') as f:
    doc = json.load(f)

# Read historical txs (JSONL format, amino JSON)
historical_txs = []
with open('$TX_EXPORT') as f:
    for line in f:
        line = line.strip()
        if line:
            historical_txs.append(json.loads(line))

# Append to existing txs
if 'app_state' not in doc:
    doc['app_state'] = {}
if 'txs' not in doc['app_state']:
    doc['app_state']['txs'] = []

existing_count = len(doc['app_state']['txs'])
doc['app_state']['txs'].extend(historical_txs)
total_count = len(doc['app_state']['txs'])

print(f'  Existing genesis txs: {existing_count}', file=sys.stderr)
print(f'  Historical txs added: {len(historical_txs)}', file=sys.stderr)
print(f'  Total txs: {total_count}', file=sys.stderr)

with open('$FINAL_GENESIS', 'w') as f:
    json.dump(doc, f, indent=2)
"

    printf "\n"
else
    printf "=== Phase 3: SKIPPED ===\n\n"
    FINAL_GENESIS="$OVERLAY_GENESIS"
fi

# =============================================================================
# Output
# =============================================================================

cp "$FINAL_GENESIS" "$OUTPUT"

HASH=$(sha256_file "$OUTPUT")
SIZE=$(du -h "$OUTPUT" | cut -f1)

printf "=== Done ===\n"
printf "  Output:         %s (%s)\n" "$OUTPUT" "$SIZE"
printf "  SHA-256:        %s\n" "$HASH"
printf "  Chain ID:       %s\n" "$CHAIN_ID"
printf "  Original Chain: %s\n" "$ORIGINAL_CHAIN_ID"
if [ -n "$HALT_HEIGHT" ]; then
    printf "  Initial Height: %d\n" "$((HALT_HEIGHT + 1))"
fi
printf "\n"
printf "Next steps:\n"
printf "  1. Share SHA-256 with other validators\n"
printf "  2. gnoland start --genesis %s\n" "$OUTPUT"
