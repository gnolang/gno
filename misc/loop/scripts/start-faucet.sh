#!/usr/bin/env sh

CHAIN_ID="${CHAIN_ID:-dev}"
RPC_URL="${RPC_URL:-http://localhost:26657}"

HOST="${HOST}:-0.0.0.0"
PORT="${PORT}:-5050"

exec gnofaucet \
    --chain-id "${CHAIN_ID}" \
    --remote "${RPC_URL}" \
    --listen-address "${HOST}:${PORT}" \
    --fund-limit 250000000ugnot \
    --send-amount 1000000000ugnot \
    --num-accounts 50 \
    --mnemonic "replace creek alley example ride access morning foot grid glory mixture hurdle pause pipe hen require tide salute music young total jaguar world dragon"
