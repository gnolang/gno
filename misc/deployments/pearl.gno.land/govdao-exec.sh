#!/usr/bin/env bash
# govdao wrapper for pearl.
# Usage: ./govdao-exec.sh [command] [args...]
export GOVDAO_LABEL="pearl"
export GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
export CHAIN_ID="${CHAIN_ID:-pearl-1}"
export REMOTE="${REMOTE:-https://rpc.pearl.testnets.gno.land}"
source "$(cd "$(dirname "$0")/../../govdao-scripts" && pwd)/govdao-wrapper.sh" "$@"
