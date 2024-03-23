#!/usr/bin/env sh

MONIKER=${MONIKER:-"gnode"}
P2P_LADDR=${P2P_LADDR:-"tcp://0.0.0.0:26656"}
RPC_LADDR=${RPC_LADDR:-"tcp://0.0.0.0:26657"}

SEEDS=${SEEDS:-""}
PERSISTENT_PEERS=${PERSISTENT_PEERS:-""}

mkdir -p testdir/data
echo '{ "height": "0", "round": "0", "step": 0 }' > testdir/data/priv_validator_state.json

gnoland -skip-start=true

sed -i "s#^moniker = \".*\"#moniker = \"${MONIKER}\"#" ./testdir/config/config.toml
sed -i "s#^laddr = \".*:26656\"#laddr = \"${P2P_LADDR}\"#" ./testdir/config/config.toml
sed -i "s#^laddr = \".*:26657\"#laddr = \"${RPC_LADDR}\"#" ./testdir/config/config.toml

sed -i "s#^seeds = \".*\"#seeds = \"${SEEDS}\"#" ./testdir/config/config.toml
sed -i "s#^persistent_peers = \".*\"#persistent_peers = \"${PERSISTENT_PEERS}\"#" ./testdir/config/config.toml

exec gnoland
