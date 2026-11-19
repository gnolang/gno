#!/bin/sh
# Shared setup for gnovm audit scripts.
# Sources the test1 key into GNOKEY_HOME if not already present.

RPC="${RPC:-http://gnoland:26657}"
CHAINID="${CHAINID:-test}"
KEY="${KEY:-test1}"
PASSWORD="${PASSWORD:-test1234}"
GNOKEY_HOME="${GNOKEY_HOME:-/tmp/gnokey}"
# Deterministic address for the test1 key (derived from the mnemonic below)
KEY_ADDR="g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

TEST1_MNEMONIC="source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"

if ! gnokey list -home "$GNOKEY_HOME" 2>/dev/null | grep -q "$KEY"; then
    printf "%s\n%s\n%s\n" "$TEST1_MNEMONIC" "$PASSWORD" "$PASSWORD" | \
        gnokey add "$KEY" -recover -insecure-password-stdin=true -home "$GNOKEY_HOME" > /dev/null 2>&1
fi
