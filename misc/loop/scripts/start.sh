#!/usr/bin/env sh

MONIKER=${MONIKER:-"the-staging-chain"}
P2P_LADDR=${P2P_LADDR:-"tcp://0.0.0.0:26656"}
RPC_LADDR=${RPC_LADDR:-"tcp://0.0.0.0:26657"}

CHAIN_ID=${CHAIN_ID:-"staging"}

GENESIS_BACKUP_FILE=${GENESIS_BACKUP_FILE:-""}
GENESIS_BALANCES_FILE=${GENESIS_BALANCES_FILE:-""}

SEEDS=${SEEDS:-""}
PERSISTENT_PEERS=${PERSISTENT_PEERS:-""}
FINAL_GENESIS_TXS_SHEET="/gnoroot/gno.land/genesis/genesis_txs.jsonl"

echo "" >> $FINAL_GENESIS_TXS_SHEET
cat "${GENESIS_BACKUP_FILE}" >> $FINAL_GENESIS_TXS_SHEET

# Reset balance file if backup file is not empty
if [ -n "$GENESIS_BALANCES_FILE" ] && [ -s "$GENESIS_BALANCES_FILE" ]; then
  cat "$GENESIS_BALANCES_FILE" /gnoroot/gno.land/genesis/genesis_balances.txt > /tmp/genesis_balances.tmp
  mv /tmp/genesis_balances.tmp /gnoroot/gno.land/genesis/genesis_balances.txt
fi

# Add FAUCET ADDRESS
[ -n "${FAUCET_ADDRESS}" ] && echo "$FAUCET_ADDRESS=10000000000000ugnot" >> /gnoroot/gno.land/genesis/genesis_balances.txt

# Initialize the secrets
gnoland secrets init

# Initialize the configuration
gnoland config init

# Set the config values
gnoland config set moniker "${MONIKER}"
gnoland config set rpc.laddr "${RPC_LADDR}"
gnoland config set p2p.laddr "${P2P_LADDR}"
gnoland config set p2p.seeds "${SEEDS}"
gnoland config set p2p.persistent_peers "${PERSISTENT_PEERS}"

# Running a lazy init will generate a fresh genesis.json, with
# the previously generated secrets. We do this to avoid CLI magic from config
# reading and piping to the gnoland genesis commands
exec gnoland start \
         --chainid="${CHAIN_ID}" \
         --genesis-txs-file="${FINAL_GENESIS_TXS_SHEET}" \
         --lazy \
         --skip-failing-genesis-txs \
         --skip-genesis-sig-verification
