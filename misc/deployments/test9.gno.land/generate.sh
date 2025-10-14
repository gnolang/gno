#!/usr/bin/env bash

# This script generates the test9 genesis.json locally, by using external
# sources for large artifacts such as genesis balances

set -e # exit on error

pullBalances () {
  local TARGET_DIR=$1
  local BALANCES_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test9/genesis_balances.txt"
  local TARGET_FILE="$TARGET_DIR/genesis_balances.txt"

  mkdir -p "$TARGET_DIR"
  wget -O "$TARGET_FILE" "$BALANCES_URL"
}

pullTxs () {
    local TARGET_DIR=$1
    local TXS_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test9/genesis_txs.jsonl"
    local TARGET_FILE="$TARGET_DIR/genesis_txs.jsonl"

    mkdir -p "$TARGET_DIR"
    wget -O "$TARGET_FILE" "$TXS_URL"
}

CHAIN_ID=test9
GENESIS_TIME=1760428800 # Tuesday, October 14th 2025 10:00 GMT+0200 (Central European Summer Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
echo "Generating fresh genesis..."
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (6)
printf "\nAdding validators...\n"

# Gno Core (2)
gnogenesis validator add -name gnocore-val-01 -power 1 -address XXX -pub-key XXX # TODO @Sergio
gnogenesis validator add -name gnocore-val-01 -power 1 -address XXX -pub-key XXX # TODO @Sergio
# Gno DevX (2)
gnogenesis validator add -name gnodevx-val-01 -power 1 -address XXX -pub-key XXX # TODO @Salvo
gnogenesis validator add -name gnodevx-val-01 -power 1 -address XXX -pub-key XXX # TODO @Salvo
# Onbloc (2)
gnogenesis validator add -name onbloc-val-01 -power 1 -address XXX -pub-key XXX # TODO @Jinseon
gnogenesis validator add -name onbloc-val-01 -power 1 -address XXX -pub-key XXX # TODO @Jinseon

# Use a temporary directory for intermediary states
TMP_DIR=./tmp-genesis
TXS_PATH=$TMP_DIR/genesis_txs.jsonl

printf "\nAdding txs (this may take a while)...\n"

pullTxs $TMP_DIR
gnogenesis txs add sheets $TXS_PATH

# Add the balances.
# Since there is a significant number of balances
# for the test9 deployment (~42MB), this balance sheet is stored
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
