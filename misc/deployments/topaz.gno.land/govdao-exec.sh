#!/usr/bin/env bash
# govdao wrapper for topaz.
# Usage: ./govdao [command] [args...]
export GOVDAO_LABEL="topaz"
export GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
export CHAIN_ID="${CHAIN_ID:-topaz-1}"
export REMOTE="${REMOTE:-https://rpc.topaz.testnets.gno.land}"
source "$(cd "$(dirname "$0")/../../govdao-scripts" && pwd)/govdao-wrapper.sh" "$@"
