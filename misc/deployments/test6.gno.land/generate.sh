#!/usr/bin/env bash

# This script generates the test6 genesis.json
# locally, by using external

set -e # exit on error

pullBalances () {
  local TARGET_DIR=$1
  # TODO update URL
  local BALANCES_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test6-genesis-balances.txt"

  mkdir -p "$TARGET_DIR"
  wget -O "$TARGET_DIR/genesis_balances.txt" "$BALANCES_URL"
}

CHAIN_ID=test6
GENESIS_TIME=1744614000 # Monday, April 14th 2025 09:00 GMT+0200 (Central European Summer Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (8)
# Gno Core (2)
gnogenesis validator add -name gnocore-val-01 -power 1 -address TODO -pub-key TODO
gnogenesis validator add -name gnocore-val-02 -power 1 -address TODO -pub-key TODO

# Gno DevX (2)
gnogenesis validator add -name devx-val-01 -power 1 -address TODO -pub-key TODO
gnogenesis validator add -name devx-val-02 -power 1 -address TODO -pub-key TODO

# AiB (2)
gnogenesis validator add -name aib-val-01 -power 1 -address TODO -pub-key TODO
gnogenesis validator add -name aib-val-02 -power 1 -address TODO -pub-key TODO

# Onbloc (2)
gnogenesis validator add -name onbloc-val-01 -power 1 -address g14cppfre9hsvu6p4scttuyu7mj082lfwxl7hvz9 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqe87d7lc0c4l4yaa8a94fucfre8882n8556l9z5220zjaaaqj7k5cl5sud
gnogenesis validator add -name onbloc-val-02 -power 1 -address g1927k3s7q9ujla04r5zy7q5m3gl84wsrart6663 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq762adl0tcvdn54d6nzqf68r9wrywn7zj87v92mk3qpr436mevpvc63wsz

# Add the transactions (all examples).
# Test1 is the deployer key for all genesis transactions, and
# it has an adequate premine amount in the balances already
gnogenesis txs add packages ../../../examples

# Add the balances.
# Since there is a significant number of balances
# for the test6 deployment (~42MB), this balance sheet is stored
# externally and fetched to generate the genesis.json
BALANCES_PATH=./tmp-genesis/genesis_balances.txt

pullBalances $BALANCES_PATH
gnogenesis balances add -balance-sheet $BALANCES_PATH

# Verify that the genesis.json is valid
gnogenesis verify -genesis-path $GENESIS_FILE

# Verify the checksum, if any
if [[ -n "$CHECKSUM" ]]; then
  ACTUAL_CHECKSUM=$(sha256sum "$GENESIS_FILE" | awk '{print $1}')

  if [[ "$ACTUAL_CHECKSUM" != "$CHECKSUM" ]]; then
    echo "❌ Genesis checksum mismatch"
    echo "Expected: $CHECKSUM"
    echo "Actual:   $ACTUAL_CHECKSUM"

    return 1
  fi

  echo "✅ Checksum verified"

  return 0
fi