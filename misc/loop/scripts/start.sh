#!/usr/bin/env sh

MONIKER=${MONIKER:-"gnode"}
P2P_LADDR=${P2P_LADDR:-"tcp://0.0.0.0:26656"}
RPC_LADDR=${RPC_LADDR:-"tcp://0.0.0.0:26657"}

CHAIN_ID=${CHAIN_ID:-"portal-loop"}

GENESIS_BACKUP_FILE=${GENESIS_BACKUP_FILE:-""}

SEEDS=${SEEDS:-""}
PERSISTENT_PEERS=${PERSISTENT_PEERS:-""}

echo "" >> /opt/gno/src/gno.land/genesis/genesis_txs.jsonl
cat ${GENESIS_BACKUP_FILE} >> /opt/gno/src/gno.land/genesis/genesis_txs.jsonl

gnoland start \
    --chainid="${CHAIN_ID}" \
    --skip-start=true \
    --skip-failing-genesis-txs

sed -i "s#^moniker = \".*\"#moniker = \"${MONIKER}\"#" ./gnoland-data/config/config.toml
sed -i "s#laddr = \".*:26656\"#laddr = \"${P2P_LADDR}\"#" ./gnoland-data/config/config.toml
sed -i "s#laddr = \".*:26657\"#laddr = \"${RPC_LADDR}\"#" ./gnoland-data/config/config.toml

sed -i "s#seeds = \".*\"#seeds = \"${SEEDS}\"#" ./gnoland-data/config/config.toml
sed -i "s#persistent_peers = \".*\"#persistent_peers = \"${PERSISTENT_PEERS}\"#" ./gnoland-data/config/config.toml

exec gnoland start --skip-failing-genesis-txs
