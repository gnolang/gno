#!/usr/bin/env bash
# govdao wrapper for test-13.
# Usage: ./govdao [command] [args...]
export GOVDAO_LABEL="test-13"
export GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
export CHAIN_ID="${CHAIN_ID:-test-13}"
export REMOTE="${REMOTE:-127.0.0.1:26657}"
source "$(cd "$(dirname "$0")/../../govdao-scripts" && pwd)/govdao-wrapper.sh" "$@"
