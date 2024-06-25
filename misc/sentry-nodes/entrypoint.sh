#!/usr/bin/env sh

MONIKER=${MONIKER:-"gnode"}
P2P_LADDR=${P2P_LADDR:-"tcp://0.0.0.0:26656"}
RPC_LADDR=${RPC_LADDR:-"tcp://0.0.0.0:26657"}

P2P_PEX=${P2P_PEX:-"true"}
P2P_PRIVATE_PEER_IDS=${P2P_PRIVATE_PEER_IDS:-""}
SEED_MODE=${SEED_MODE:-"false"}

CHAIN_ID=${CHAIN_ID:-"dev"}

SEEDS=${SEEDS:-""}
PERSISTENT_PEERS=${PERSISTENT_PEERS:-""}

# echo '{}' > gnoland-data/secrets/priv_validator_state.json

gnoland config init

# Set the config values

gnoland config set moniker       "${MONIKER}"
gnoland config set rpc.laddr     "${RPC_LADDR}"
gnoland config set p2p.laddr     "${P2P_LADDR}"

gnoland config set p2p.pex              "${P2P_PEX}"
gnoland config set p2p.private_peer_ids "${P2P_PRIVATE_PEER_IDS}"

gnoland config set p2p.seed_mode "${SEED_MODE}"

gnoland config set p2p.seeds     "${SEEDS}"
gnoland config set p2p.persistent_peers "${PERSISTENT_PEERS}"

exec gnoland start --genesis="./gnoland-data/genesis.json" --log-level=info
