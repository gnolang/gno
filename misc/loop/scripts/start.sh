#!/usr/bin/env sh

MONIKER=${MONIKER:-"gnode"}
P2P_LADDR=${P2P_LADDR:-"tcp://0.0.0.0:26656"}
RPC_LADDR=${RPC_LADDR:-"tcp://0.0.0.0:26657"}

CHAIN_ID=${CHAIN_ID:-"portal-loop"}

GENESIS_BACKUP_FILE=${GENESIS_BACKUP_FILE:-""}

SEEDS=${SEEDS:-""}
PERSISTENT_PEERS=${PERSISTENT_PEERS:-""}

echo "" >> /gnoroot/gno.land/genesis/genesis_txs.jsonl
cat ${GENESIS_BACKUP_FILE} >> /gnoroot/gno.land/genesis/genesis_txs.jsonl

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
         --lazy \
         --skip-failing-genesis-txs
