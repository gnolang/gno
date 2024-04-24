#!/usr/bin/env sh
# This script is used by portalloopd to actually launch the gno.land node.
# It sets up the gno.land node's config with the given environment variables,
# and starts the node.

MONIKER=${MONIKER:-"gnode"}
P2P_LADDR=${P2P_LADDR:-"tcp://0.0.0.0:26656"}
RPC_LADDR=${RPC_LADDR:-"tcp://0.0.0.0:26657"}

CHAIN_ID=${CHAIN_ID:-"portal-loop"}

GENESIS_BACKUP_FILE=${GENESIS_BACKUP_FILE:-""}

SEEDS=${SEEDS:-""}
PERSISTENT_PEERS=${PERSISTENT_PEERS:-""}

echo "" >> /opt/gno/src/gno.land/genesis/genesis_txs.jsonl
cat ${GENESIS_BACKUP_FILE} >> /opt/gno/src/gno.land/genesis/genesis_txs.jsonl

gnoland config init -config-path="./testdir/config/config.toml"

gnoland config set -config-path="./testdir/config/config.toml" moniker "${MONIKER}"
gnoland config set -config-path="./testdir/config/config.toml" p2p.laddr "${P2P_LADDR}"
gnoland config set -config-path="./testdir/config/config.toml" rpc.laddr "${RPC_LADDR}"

gnoland config set -config-path="./testdir/config/config.toml" p2p.seeds "${SEEDS}"
gnoland config set -config-path="./testdir/config/config.toml" p2p.persistent_peers "${PERSISTENT_PEERS}"

exec gnoland start --skip-failing-genesis-txs --chainid="${CHAIN_ID}"
