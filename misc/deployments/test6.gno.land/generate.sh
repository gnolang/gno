#!/usr/bin/env bash

# This script generates the test6 genesis.json locally, by using external
# sources for large artifacts such as genesis balances

set -e # exit on error

pullBalances () {
  local TARGET_DIR=$1
  local BALANCES_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test6/genesis_balances.txt"
  local TARGET_FILE="$TARGET_DIR/genesis_balances.txt"

  mkdir -p "$TARGET_DIR"
  wget -O "$TARGET_FILE" "$BALANCES_URL"
}

CHAIN_ID=test6
GENESIS_TIME=1744614000 # Monday, April 14th 2025 09:00 GMT+0200 (Central European Summer Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
echo "Generating fresh genesis..."
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (8)
printf "\nAdding validators...\n"

# Gno Core (2)
gnogenesis validator add -name gnocore-val-01 -power 1 -address g1wsa9j6nel8ltt6q2lmf78585ymyfh5nsvhaxa3 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqma223maxmnw4f42kfqppvgyn8dr8wu7mhtdm6lcq64303a3vlln8xdmms
gnogenesis validator add -name gnocore-val-02 -power 1 -address g13762rd7y8s7jcc6uc4lyxv269hguchhpyzaamt -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq5ndww8w6qrxgfdeastcx2lsuuk5r8w9jckkgevylq6duw59d54n935fq2

# Gno DevX (2)
# gnogenesis validator add -name devx-val-01 -power 1 -address TODO -pub-key TODO
# gnogenesis validator add -name devx-val-02 -power 1 -address TODO -pub-key TODO

# AiB (2)
gnogenesis validator add -name aib-val-01 -power 1 -address g12yvv8pl5s20suxyd30g7ychqenamtlhctgfu90 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zplgfkp8609ghdh20w6newh40f9tz7ussw2zylq23ca0tjda3csztm242ft
gnogenesis validator add -name aib-val-02 -power 1 -address g1p3lyk676gludkk6hqceem58c6xgnpsld45s4v9 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpf65yj5xh8y9qux89skvve77w7hytfcfey92zlvx56ruugqvk9eepk73fg

# Onbloc (2)
gnogenesis validator add -name onbloc-val-01 -power 1 -address g14cppfre9hsvu6p4scttuyu7mj082lfwxl7hvz9 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqe87d7lc0c4l4yaa8a94fucfre8882n8556l9z5220zjaaaqj7k5cl5sud
gnogenesis validator add -name onbloc-val-02 -power 1 -address g1927k3s7q9ujla04r5zy7q5m3gl84wsrart6663 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq762adl0tcvdn54d6nzqf68r9wrywn7zj87v92mk3qpr436mevpvc63wsz


printf "\nAdding txs (this may take a while)...\n"
# Add the transactions (all examples).
# Test1 is the deployer key for all genesis transactions, and
# it has an adequate premine amount in the balances already
gnogenesis txs add packages ../../../examples

# Add the balances.
# Since there is a significant number of balances
# for the test6 deployment (~42MB), this balance sheet is stored
# externally and fetched to generate the genesis.json
BALANCES_DIR=./tmp-genesis
BALANCES_PATH=$BALANCES_DIR/genesis_balances.txt

printf "\nAdding balances...\n"

pullBalances $BALANCES_DIR
gnogenesis balances add -balance-sheet $BALANCES_PATH

rm -rf $BALANCES_DIR

# Verify that the genesis.json is valid
printf "\Verifying genesis.json...\n"
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
