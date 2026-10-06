#!/usr/bin/env bash
# govdao wrapper for sapphire.
# Usage: ./govdao-exec.sh [command] [args...]
export GOVDAO_LABEL="sapphire"
export GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
export CHAIN_ID="${CHAIN_ID:-sapphire-1}"
export REMOTE="${REMOTE:-https://rpc.sapphire.testnets.gno.land}"
source "$(cd "$(dirname "$0")/../../govdao-scripts" && pwd)/govdao-wrapper.sh" "$@"
