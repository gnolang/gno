#!/usr/bin/env bash

# This script generates the test8 genesis.json locally, by using external
# sources for large artifacts such as genesis balances

set -e # exit on error

pullBalances () {
  local TARGET_DIR=$1
  local BALANCES_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test8/genesis_balances.txt"
  local TARGET_FILE="$TARGET_DIR/genesis_balances.txt"

  mkdir -p "$TARGET_DIR"
  wget -O "$TARGET_FILE" "$BALANCES_URL"
}

pullTxs () {
    local TARGET_DIR=$1
    local TXS_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test8/genesis_txs.jsonl"
    local TARGET_FILE="$TARGET_DIR/genesis_txs.jsonl"

    mkdir -p "$TARGET_DIR"
    wget -O "$TARGET_FILE" "$TXS_URL"
}

CHAIN_ID=test8
GENESIS_TIME=1757055600 # Friday, September 5th 2025 09:00 GMT+0200 (Central European Summer Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
echo "Generating fresh genesis..."
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (4)
printf "\nAdding validators...\n"

# Gno Core (2) + Onbloc (2)
gnogenesis validator add -name gnocore-val-01 -power 1 -address g1cc9x4xyf3n3jygf3fvz462k9w0qaklda39mh2c -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqgd6ya9uc5auw4a2y8ms4jyqvdvu453jruseclyz6xnk5qltxvvnwjtkrs
gnogenesis validator add -name gnocore-val-02 -power 1 -address g1854g5lq87x5f75etzmg2wcd2mdmvj96hr469sa -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpkljjay9cnqmtdcmndxd355ggga8s0tmmu0g5fdplu592wju6spfzngswj
gnogenesis validator add -name gnocore-val-03 -power 1 -address g15zkeyz2gwrjluqj6eremllrh6nx7mt4tlz8f32 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpkhggjkjpvuuf4vmdv5lws9f3c22j6qkys3djt44dsyqwqc9padxlkhfd0
gnogenesis validator add -name gnocore-val-04 -power 1 -address g1v7wl7qlakzku5mrafmgntfuvd7xjrluhuhwewp -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqvl9cfzralqscxunw7cus2dmjwveskwcldjwnpr29zps6d5tlv5e2q2443

# Use a temporary directory for intermediary states
TMP_DIR=./tmp-genesis
TXS_PATH=$TMP_DIR/genesis_txs.jsonl

printf "\nAdding txs (this may take a while)...\n"

pullTxs $TMP_DIR
gnogenesis txs add sheets $TXS_PATH

# Add the balances.
# Since there is a significant number of balances
# for the test8 deployment (~42MB), this balance sheet is stored
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
