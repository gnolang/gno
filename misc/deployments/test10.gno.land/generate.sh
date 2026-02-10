#!/usr/bin/env bash

# This script generates the test10 genesis.json locally, by using external
# sources for large artifacts such as genesis balances

set -e # exit on error

pullBalances () {
  local TARGET_DIR=$1
  local BALANCES_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test10/genesis_balances.txt"
  local TARGET_FILE="$TARGET_DIR/genesis_balances.txt"

  mkdir -p "$TARGET_DIR"
  wget -O "$TARGET_FILE" "$BALANCES_URL"
}

pullTxs () {
    local TARGET_DIR=$1
    local TXS_URL="https://gno-testnets-genesis.s3.eu-central-1.amazonaws.com/test10/genesis_txs.jsonl"
    local TARGET_FILE="$TARGET_DIR/genesis_txs.jsonl"

    mkdir -p "$TARGET_DIR"
    wget -O "$TARGET_FILE" "$TXS_URL"
}

CHAIN_ID=test10
GENESIS_TIME=1766044800 # Thursday, December 18th 2025 09:00 GMT+0100 (Central European Standard Time)
GENESIS_FILE=genesis.json

# Generate a fresh genesis.json
echo "Generating fresh genesis..."
gnogenesis generate -chain-id $CHAIN_ID -genesis-time $GENESIS_TIME -output-path $GENESIS_FILE

# Add the initial validators (4)
printf "\nAdding validators...\n"

# Gno Core (2)
gnogenesis validator add -name gnocore-val-01 -power 1 -address g1gmg597aa85gk6u3wz3aluyxmctgfq9ld2fda7w -pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqvsdv9yhul20sq9zfk08r3u72ysqdntu59xcez0ju2ydjguya9fusuc8h27
gnogenesis validator add -name gnocore-val-02 -power 1 -address g1y8uw54dytc3twhv0vr4h8erghh7vxkczxvz9x0 -pub-key gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq06uuq3asdxgs5f438y863wkly5flazwtaz83ldvh33pmdecd0s0jyndfnl

# Onbloc (2)
gnogenesis validator add -name onbloc-val-01 -power 1 -address g1wmaglcam7xq3kwvrks6ysyeutf0jc877f4q3nl -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpelzea0ep9dclr83vygetymcs5uekam9j53t2sgcefumxth2wagquu3vrk
gnogenesis validator add -name onbloc-val-02 -power 1 -address g1quanjrv29f3zvy8cnsd7s3wjcnhmv4v4n3knhm -pub-key gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqz3ewj7qudswvtsex2l73g42v5tnkyguxcgft8tdg3mrpzwzzn4lde0vag

# Use a temporary directory for intermediary states
TMP_DIR=./tmp-genesis
TXS_PATH=$TMP_DIR/genesis_txs.jsonl

printf "\nAdding txs (this may take a while)...\n"

pullTxs $TMP_DIR
gnogenesis txs add sheets $TXS_PATH

# Add the balances.
# Since there is a significant number of balances
# for the test10 deployment (~42MB), this balance sheet is stored
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
