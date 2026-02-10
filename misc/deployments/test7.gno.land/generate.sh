#!/usr/bin/env bash

# This script generates the test7 genesis.json locally, by using external
# sources for large artifacts such as genesis balances

set -e # exit on error

pullBalances () {
  local TARGET_DIR=$1
  local BALANCES_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test7/genesis_balances.txt"
  local TARGET_FILE="$TARGET_DIR/genesis_balances.txt"

  mkdir -p "$TARGET_DIR"
  wget -O "$TARGET_FILE" "$BALANCES_URL"
}

pullTxs () {
    local TARGET_DIR=$1
    local TXS_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test7/genesis_txs.jsonl"
    local TARGET_FILE="$TARGET_DIR/genesis_txs.jsonl"

    mkdir -p "$TARGET_DIR"
    wget -O "$TARGET_FILE" "$TXS_URL"
}

CHAIN_ID=test7.2
GENESIS_TIME=1753862400 # Wednesday, July 30th 2025 10:00 GMT+0200 (Central European Summer Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
echo "Generating fresh genesis..."
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (8)
printf "\nAdding validators...\n"

# Gno Core (4)
gnogenesis validator add -name gnocore-val-01 -power 1 -address g1mt5d6l56tf8r3u8mehnhgp54mqlx3lu5qdysum -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpv5dep4jymxp9p3xkd5e3cdk63wqygw00nsm402nxh593rkuan75gkr7e6
gnogenesis validator add -name gnocore-val-02 -power 1 -address g1mmrdx5j4v878uqttp44t96d3rlw8rls8fgze24 -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpvklu0t8l37fh4lf8kap8jcrsk3akkvm5nsu6m2v74agp45xt9pd7h5ana
gnogenesis validator add -name gnocore-val-03 -power 1 -address g16wh3t370fctrukvvdslsz9uc76tfpy0ggrm5hr -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp0768mk3fg7dkprgkl5twxzhn2hlef27mzt46qea8z4qv0ltwstuy9qk62
gnogenesis validator add -name gnocore-val-04 -power 1 -address g1y7d86mffahwy7c8s4j0nvgwhkg4n4tl7ek4w8z -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq82uuwf5yvrfpyhp5zh5hspz7wvee4jpsh37muczqrpatuxr3vpwyhjdsy

# Use a temporary directory for intermediary states
TMP_DIR=./tmp-genesis
TXS_PATH=$TMP_DIR/genesis_txs.jsonl

printf "\nAdding txs (this may take a while)...\n"

pullTxs $TMP_DIR
gnogenesis txs add sheets $TXS_PATH

# Add the balances.
# Since there is a significant number of balances
# for the test7 deployment (~42MB), this balance sheet is stored
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
