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

CHAIN_ID=test9.1
GENESIS_TIME=1762329600 # Wednesday, November 5th 2025 09:00 GMT+0200 (Central European Summer Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
echo "Generating fresh genesis..."
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (6)
printf "\nAdding validators...\n"

# Gno Core (2)
gnogenesis validator add -name gnocore-val-01 -power 1 -address g1ek7ftha29qv4ahtv7jzpc0d57lqy7ynzklht7t -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq7kdacwkqquf7j9ywd6c3yaj4vsytkan526knd7nt2z5v4q2mgc6hfkfyu
gnogenesis validator add -name gnocore-val-02 -power 1 -address g1xwlzxpuh4v3l9fjrl4cq2ylzpk7mhj0598xdy7 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpnxghmy95x979vyw43073prpzqah9k9sez2mxsfkyfxge02hqczhqy4r4j

# Gno DevX (2)
gnogenesis validator add -name gnodevx-val-01 -power 1 -address g1zeawgdp4nh6j6phmnsjmssqrz4ukruh43anckl -pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqwhdpj834lxwenxja6tanrz2rn3a4rpcxxqxhh23f77dsftyhqrgxpghjwm
gnogenesis validator add -name gnodevx-val-02 -power 1 -address g1uajmkwz6w7juqxhu64qecfm6kr4r8us70culre -pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqfc9agrkxsm7d4ry6zz823f7ulkkc3u3nqv6avz99tm45nmfj4ug54c38c0

# Onbloc (2)
gnogenesis validator add -name onbloc-val-01 -power 1 -address g1kntcjkfplj0z44phajajwqkx5q4ry5yaft5q2h -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqxkeawgn93f8jm7wumknjwysg5mm3shyeyh8fky6pwytpcf5uyqk4vwshh
gnogenesis validator add -name onbloc-val-02 -power 1 -address g1j306jcl4qyhgjw78shl3ajp88vmvdcf7m7ntm2 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqvpv84g5n8zz7k2m2yenuv73e9r77zxzxzrucj064gzykj2puyvszw8kxc

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
