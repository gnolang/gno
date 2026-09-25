#!/usr/bin/env bash
# misc/hf-glue/scripts/init-cluster.sh
#
# Initialise a multi-node gnoland cluster for the testbed (default: 2 nodes,
# hardcoded to 2 in docker-compose.cluster.yml — bump NODES and edit the
# compose file if you want more).
#
# Layout:
#   $OUT/cluster/node0/{secrets,config,genesis.json}
#   $OUT/cluster/node1/...
#
# Secrets are independent per-node. The genesis is shared — rewritten once
# to contain all N validators — and copied into each node's home. Each
# config.toml is patched with `persistent_peers` so the nodes dial each
# other over the docker-compose network (service names: node0, node1, ...).
#
# Env:
#   NODES              number of nodes (default: 2)
#   VALIDATOR_NAME     base name for validators (suffixed -N)
#   OUT, REPO          absolute paths
set -euo pipefail

: "${OUT:?OUT is required}"
: "${REPO:?REPO is required}"
: "${VALIDATOR_NAME:=hf-glue-cluster}"
NODES="${NODES:-2}"

GENESIS="$OUT/genesis.json"
CLUSTER_DIR="$OUT/cluster"

if [[ ! -f "$GENESIS" ]]; then
  echo "missing $GENESIS — run 'make fetch' first" >&2
  exit 1
fi

echo "── init $NODES-validator cluster ────────────────────────────"
mkdir -p "$CLUSTER_DIR"

# ---- 1. generate per-node secrets ----------------------------------------
PRIV_KEYS=()
for ((i=0; i<NODES; i++)); do
  HOME_DIR="$CLUSTER_DIR/node$i"
  SECRETS_DIR="$HOME_DIR/secrets"
  PV_KEY="$SECRETS_DIR/priv_validator_key.json"
  mkdir -p "$HOME_DIR"
  if [[ -f "$PV_KEY" ]]; then
    echo "  node$i: secrets present at $SECRETS_DIR — reusing"
  else
    echo "  node$i: generating secrets in $SECRETS_DIR"
    mkdir -p "$SECRETS_DIR"
    go run -C "$REPO" ./gno.land/cmd/gnoland secrets init --data-dir "$SECRETS_DIR" >/dev/null
  fi
  PRIV_KEYS+=(--priv-key "$PV_KEY")
done

# ---- 2. rewrite the shared genesis with ALL validators -------------------
echo ""
echo "  rewriting validator set in $GENESIS ($NODES validators)..."
go run -C "$REPO/misc/hf-glue/fixvalidator" . \
  "${PRIV_KEYS[@]}" \
  --genesis "$GENESIS" \
  --name "$VALIDATOR_NAME" \
  --power 10

# ---- 3. collect node IDs for persistent_peers ----------------------------
declare -a NODE_IDS
for ((i=0; i<NODES; i++)); do
  ID=$(go run -C "$REPO" ./gno.land/cmd/gnoland secrets get node_id.id --raw \
       -data-dir "$CLUSTER_DIR/node$i/secrets" | tr -d '[:space:]')
  NODE_IDS[$i]="$ID"
  echo "  node$i id: $ID"
done

# ---- 4. per-node config.toml --------------------------------------------
# Inside the docker-compose network, each service is reachable by its name
# (node0, node1, ...) on the container-internal p2p port 26656.
P2P_PORT_INTERNAL=26656

for ((i=0; i<NODES; i++)); do
  HOME_DIR="$CLUSTER_DIR/node$i"
  CONFIG_DIR="$HOME_DIR/config"
  mkdir -p "$CONFIG_DIR"
  go run -C "$REPO" ./gno.land/cmd/gnoland config init -config-path "$CONFIG_DIR/config.toml" -force >/dev/null

  # build persistent_peers list = every other node
  PEERS=""
  for ((j=0; j<NODES; j++)); do
    [[ $i -eq $j ]] && continue
    entry="${NODE_IDS[$j]}@node$j:$P2P_PORT_INTERNAL"
    PEERS="${PEERS:+$PEERS,}$entry"
  done

  sed -i.bak \
    -e 's|tcp://127.0.0.1:26657|tcp://0.0.0.0:26657|' \
    -e 's|tcp://127.0.0.1:26656|tcp://0.0.0.0:26656|' \
    -e "s|^persistent_peers = .*|persistent_peers = \"$PEERS\"|" \
    -e "s|^moniker = .*|moniker = \"$VALIDATOR_NAME-$i\"|" \
    "$CONFIG_DIR/config.toml"
  rm -f "$CONFIG_DIR/config.toml.bak"

  # stage genesis
  cp "$GENESIS" "$HOME_DIR/genesis.json"

  echo "  node$i: persistent_peers=$PEERS"
done

echo ""
echo "done — cluster homes ready under $CLUSTER_DIR"
