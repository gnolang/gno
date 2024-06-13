#!/usr/bin/env sh

set -ex

MONIKER=${MONIKER:-"gnode"}
P2P_LADDR=${P2P_LADDR:-"tcp://0.0.0.0:26656"}
RPC_LADDR=${RPC_LADDR:-"tcp://0.0.0.0:26657"}

CHAIN_ID=${CHAIN_ID:-"staging"}

gnoland config  init
gnoland secrets init

gnoland config set moniker "${MONIKER}"
gnoland config set rpc.laddr "${RPC_LADDR}"
gnoland config set p2p.laddr "${P2P_LADDR}"

exec gnoland start \
    --skip-failing-genesis-txs \
    --chainid="${CHAIN_ID}" \
    --lazy
