#!/usr/bin/env bash
# Generate gnoland-1 hardfork genesis.
#
# This script is the gnoland-1-specific wrapper around the generic
# `gnogenesis fork generate` tool. It sets the chain-specific defaults
# (chain IDs, halt height, overlay directory) and delegates to
# `gnogenesis fork generate`.
#
# Usage:
#   ./generate-genesis.sh [SOURCE] [OPTIONS]
#
# SOURCE (positional, overrides DEFAULT_SOURCE):
#   http://...      RPC of the running or recently-halted gnoland1 node
#   /path/to/dir    local node data directory (stopped node)
#   /path/to/file   exported genesis.json or txs.jsonl
#
# OPTIONS:
#   --halt-height N   block height at which gnoland1 was halted (auto-detect if absent)
#   --output PATH     output genesis file (default: genesis.json)
#   --skip-txs        skip tx export — only copy genesis structure (fast preview)
#   --debug           print every command being run
#
# Examples:
#   ./generate-genesis.sh                              # use DEFAULT_SOURCE from this file
#   ./generate-genesis.sh http://rpc.gno.land:26657   # use production RPC
#   ./generate-genesis.sh /var/lib/gnoland             # use stopped node data dir
#   ./generate-genesis.sh --skip-txs                   # quick preview, no tx download

set -euo pipefail

# =============================================================================
# gnoland-1 specific configuration — update before each hardfork.
# =============================================================================

CHAIN_ID="gnoland-1"
ORIGINAL_CHAIN_ID="gnoland1"

# Default source: the running gnoland1 RPC endpoint.
# Override by passing the source as the first positional argument.
DEFAULT_SOURCE="http://rpc.gno.land:26657"

# Halt height: set this when the coordinated halt height is announced.
# Leave empty to auto-detect from the source (uses latest block height).
HALT_HEIGHT="${HALT_HEIGHT:-}"

# =============================================================================
# Internals — no need to edit below.
# =============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
GNOGENESIS_DIR="$REPO_ROOT/contribs/gnogenesis"
OVERLAY_DIR="$SCRIPT_DIR/overlay"
OUTPUT="${OUTPUT:-$SCRIPT_DIR/genesis.json}"

DEBUG=false
EXTRA_ARGS=()
# Initialize with set -u compatibility
SOURCE=""
SKIP_TXS=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --halt-height)    HALT_HEIGHT="$2"; shift 2 ;;
        --output)         OUTPUT="$2"; shift 2 ;;
        --skip-txs)       SKIP_TXS=true; shift ;;
        --debug)          DEBUG=true; shift ;;
        -h|--help)
            sed -n '2,/^set/{ /^set/d; s/^# //; p }' "$0"
            exit 0
            ;;
        -*)
            EXTRA_ARGS+=("$1")
            shift
            ;;
        *)
            # First positional arg is the source override
            if [[ -z "$SOURCE" ]]; then
                SOURCE="$1"
            else
                echo "ERROR: unexpected argument: $1" >&2
                exit 1
            fi
            shift
            ;;
    esac
done

# Use default source if not specified
SOURCE="${SOURCE:-$DEFAULT_SOURCE}"

run() {
    if [ "$DEBUG" = true ]; then
        printf "  \033[2m\$ %s\033[0m\n" "$*" >&2
    fi
    "$@"
}

# Build the gnogenesis binary if not already available
HARDFORK_BIN=""
if command -v gnogenesis >/dev/null 2>&1; then
    HARDFORK_BIN="gnogenesis"
else
    WORK_BIN="$SCRIPT_DIR/genesis-work/bin/gnogenesis"
    if [[ ! -x "$WORK_BIN" ]]; then
        printf "Building gnogenesis tool...\n"
        mkdir -p "$(dirname "$WORK_BIN")"
        run go build -C "$GNOGENESIS_DIR" -o "$WORK_BIN" .
        printf "  ok\n"
    fi
    HARDFORK_BIN="$WORK_BIN"
fi

# Build `gnogenesis fork generate` command
CMD_ARGS=(
    fork
    generate
    --source "$SOURCE"
    --chain-id "$CHAIN_ID"
    --original-chain-id "$ORIGINAL_CHAIN_ID"
    --output "$OUTPUT"
)

[ -n "$HALT_HEIGHT" ] && CMD_ARGS+=(--halt-height "$HALT_HEIGHT")
[ "$SKIP_TXS" = true ] && CMD_ARGS+=(--skip-txs)
[ -d "$OVERLAY_DIR" ] && CMD_ARGS+=(--overlay-dir "$OVERLAY_DIR")

[[ ${#EXTRA_ARGS[@]} -gt 0 ]] && CMD_ARGS+=("${EXTRA_ARGS[@]}")

run "$HARDFORK_BIN" "${CMD_ARGS[@]}"

# Print sha256 for cross-validator coordination
if [[ -f "$OUTPUT" ]]; then
    if command -v sha256sum >/dev/null 2>&1; then
        SHA256=$(sha256sum "$OUTPUT" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        SHA256=$(shasum -a 256 "$OUTPUT" | cut -d' ' -f1)
    fi
    printf "\nsha256: %s\n" "${SHA256:-<sha256 tool not found>}"
fi
