#!/usr/bin/env bash
# govdao wrapper for mainnet.
# Usage: ./govdao-exec.sh [command] [args...]
export GOVDAO_LABEL="mainnet"
export GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
export CHAIN_ID="${CHAIN_ID:-mainnet-1}"
export REMOTE="${REMOTE:-https://rpc.mainnet.testnets.gno.land}"
source "$(cd "$(dirname "$0")/../../govdao-scripts" && pwd)/govdao-wrapper.sh" "$@"
