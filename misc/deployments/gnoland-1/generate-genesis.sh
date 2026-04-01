#!/usr/bin/env bash
# Generate gnoland-1 genesis from a running gnoland1 node
#
# Prerequisites:
# - tx-archive binary installed (github.com/gnolang/tx-archive)
# - Access to a gnoland1 node RPC
#
# Usage:
#   ./generate-genesis.sh [--rpc URL] [--halt-height HEIGHT] [--output PATH]

set -euo pipefail

# Defaults
RPC_URL="${RPC_URL:-https://rpc.gnoland1.gno.land:443}"
HALT_HEIGHT="${HALT_HEIGHT:-}"
OUTPUT="${OUTPUT:-genesis.json}"
CHAIN_ID="gnoland-1"
ORIGINAL_CHAIN_ID="gnoland1"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Parse flags
while [[ $# -gt 0 ]]; do
    case $1 in
        --rpc)
            RPC_URL="$2"
            shift 2
            ;;
        --halt-height)
            HALT_HEIGHT="$2"
            shift 2
            ;;
        --output)
            OUTPUT="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [--rpc URL] [--halt-height HEIGHT] [--output PATH]"
            echo ""
            echo "Options:"
            echo "  --rpc URL           RPC endpoint (default: $RPC_URL)"
            echo "  --halt-height N     Export txs up to this block height"
            echo "  --output PATH       Output genesis file (default: genesis.json)"
            echo ""
            echo "Environment variables:"
            echo "  RPC_URL             Same as --rpc"
            echo "  HALT_HEIGHT         Same as --halt-height"
            echo "  OUTPUT              Same as --output"
            exit 0
            ;;
        *)
            echo "Unknown option: $1" >&2
            exit 1
            ;;
    esac
done

echo "=== Configuration ==="
echo "RPC URL:           $RPC_URL"
echo "Halt height:       ${HALT_HEIGHT:-<latest>}"
echo "Output:            $OUTPUT"
echo "Chain ID:          $CHAIN_ID"
echo "Original Chain ID: $ORIGINAL_CHAIN_ID"
echo ""

echo "=== Step 1: Export transactions from gnoland1 ==="
TX_EXPORT="$(mktemp)"
trap 'rm -f "$TX_EXPORT"' EXIT

tx-archive backup \
    --remote "$RPC_URL" \
    --output-path "$TX_EXPORT" \
    --from-block 1 \
    ${HALT_HEIGHT:+--to-block "$HALT_HEIGHT"} \
    --skip-failed-txs

echo "Exported txs to $TX_EXPORT"
echo "Total txs: $(wc -l < "$TX_EXPORT")"
echo ""

echo "=== Step 2: Assemble genesis ==="
# Use tx-archive genesis-assemble to create genesis.json
tx-archive genesis-assemble \
    --input "$TX_EXPORT" \
    --output "$OUTPUT" \
    --chain-id "$CHAIN_ID" \
    --original-chain-id "$ORIGINAL_CHAIN_ID"

echo ""
echo "=== Step 3: Verify ==="
echo "Genesis written to: $OUTPUT"
echo "Chain ID: $CHAIN_ID"
echo "Original Chain ID: $ORIGINAL_CHAIN_ID"
echo "SHA-256: $(sha256sum "$OUTPUT" | cut -d' ' -f1)"
# TODO: add genesis hash verification

echo ""
echo "=== Done ==="
echo "Next steps:"
echo "  1. Share SHA-256 of $OUTPUT with other validators"
echo "  2. Start node: gnoland start --genesis $OUTPUT"
