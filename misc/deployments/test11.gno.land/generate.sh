#!/usr/bin/env bash

# This script generates the test11 genesis.json locally, by using external
# sources for large artifacts such as genesis balances

set -e # exit on error

pullBalances () {
  local TARGET_DIR=$1
  local BALANCES_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test11/genesis_balances.txt"
  local TARGET_FILE="$TARGET_DIR/genesis_balances.txt"

  mkdir -p "$TARGET_DIR"
  wget -O "$TARGET_FILE" "$BALANCES_URL"
}

pullTxs () {
    local TARGET_DIR=$1
    local TXS_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test11/genesis_txs.jsonl"
    local TARGET_FILE="$TARGET_DIR/genesis_txs.jsonl"

    mkdir -p "$TARGET_DIR"
    wget -O "$TARGET_FILE" "$TXS_URL"
}

CHAIN_ID=test11
GENESIS_TIME=1770883200 # Thursday, February 12th 2026 09:00 GMT+0100 (Central European Standard Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
echo "Generating fresh genesis..."
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (2)
printf "\nAdding validators...\n"

# Gno Core (2)
gnogenesis validator add -name gnocore-val-01 -power 1 -address g1e8j30772p4fvjd27eqq0e50vs74ekst50f5hz6 -pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqdw9mje8pwpd8p6pu3h0ke9mxp3s7l5vr6jtsffyd9wazpgxufwrssj367n
gnogenesis validator add -name gnocore-val-02 -power 1 -address g18k5nz67cg03lrdkwnf9qmdfzf0ptaq7lxefe4y -pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0mxx3hhmy5hsl9kv8ne2ch46l8uk3kpnwc8kr900066ejz3zlhqcj6wjna

# Use a temporary directory for intermediary states
TMP_DIR=./tmp-genesis
TXS_PATH=$TMP_DIR/genesis_txs.jsonl

printf "\nAdding txs (this may take a while)...\n"

pullTxs $TMP_DIR
gnogenesis txs add sheets $TXS_PATH

# Add the balances.
# Since there is a significant number of balances
# for the test11 deployment (~42MB), this balance sheet is stored
# externally and fetched to generate the genesis.json
BALANCES_PATH=$TMP_DIR/genesis_balances.txt

printf "\nAdding balances...\n"

pullBalances $TMP_DIR
gnogenesis balances add -balance-sheet $BALANCES_PATH

# Cleanup
rm -rf $TMP_DIR

# Update the whitelisted addresses (NT + the faucet)
printf "\nUpdating whitelisted addresses...\n"

gnogenesis params set auth.unrestricted_addrs "g148583t5x66zs6p90ehad6l4qefeyaf54s69wql,g1manfred47kzduec920z88wfr64ylksmdcedlf5"

# Update the restricted denoms (enable token locking)
printf "\nEnabling token locking...\n"

gnogenesis params set bank.restricted_denoms "ugnot"

# Verify that the genesis.json is valid
printf "\nVerifying genesis.json...\n"
gnogenesis verify -genesis-path $GENESIS_FILE
