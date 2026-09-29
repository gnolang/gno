#!/usr/bin/env bash
# govdao wrapper for mainnet.
# Usage: ./govdao-exec.sh [command] [args...]
export GOVDAO_LABEL="mainnet"
export GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
export CHAIN_ID="${CHAIN_ID:-gnoland-1}" # keep in sync with CHAIN_ID in gen-genesis.sh
export REMOTE="${REMOTE:-https://rpc.gno.land}"
source "$(cd "$(dirname "$0")/../../govdao-scripts" && pwd)/govdao-wrapper.sh" "$@"
