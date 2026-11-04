#!/usr/bin/env bash
# govdao wrapper for onyx.
# Usage: ./govdao-exec.sh [command] [args...]
export GOVDAO_LABEL="onyx"
export GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
export CHAIN_ID="${CHAIN_ID:-onyx-1}" # keep in sync with CHAIN_ID in gen-genesis.sh
export REMOTE="${REMOTE:-https://rpc.onyx.testnets.gno.land}"
source "$(cd "$(dirname "$0")/../../govdao-scripts" && pwd)/govdao-wrapper.sh" "$@"
