#!/usr/bin/env bash
# Initialise gnoland-home for the testbed:
#   - run `gnoland secrets init` to generate the single validator identity
#   - rewrite validators in out/genesis.json so it contains ONLY that key
#
# The docker container mounts $OUT/gnoland-home/ as its --data-dir, so the
# node boots with the key generated here.
#
# Inputs (env):
#   VALIDATOR_NAME      name baked into the genesis validator entry
#   OUT                 output directory (absolute)
#   REPO                repo root (absolute)
set -euo pipefail

: "${VALIDATOR_NAME:?VALIDATOR_NAME is required}"
: "${OUT:?OUT is required}"
: "${REPO:?REPO is required}"

GENESIS="$OUT/genesis.json"
HOME_DIR="$OUT/gnoland-home"
SECRETS_DIR="$HOME_DIR/secrets"
PV_KEY="$SECRETS_DIR/priv_validator_key.json"

if [[ ! -f "$GENESIS" ]]; then
  echo "missing $GENESIS — run 'make fetch' first" >&2
  exit 1
fi

echo "── init single-validator node ───────────────────────────────"
mkdir -p "$HOME_DIR"

# ---- 1. generate validator secrets if not already present ----
if [[ -f "$PV_KEY" ]]; then
  echo "  secrets already present at $SECRETS_DIR — reusing"
else
  echo "  generating secrets in $SECRETS_DIR"
  mkdir -p "$SECRETS_DIR"
  go run -C "$REPO" ./gno.land/cmd/gnoland secrets init --data-dir "$SECRETS_DIR"
fi

# ---- 2. rewrite validator set in the genesis to a single entry ----
echo ""
echo "  rewriting validator set in genesis..."
go run -C "$REPO/misc/hf-glue/fixvalidator" . \
  --priv-key "$PV_KEY" \
  --genesis "$GENESIS" \
  --name "$VALIDATOR_NAME" \
  --power 10

# ---- 3. write config.toml so RPC binds to 0.0.0.0 (accessible from host) ----
CONFIG_DIR="$HOME_DIR/config"
mkdir -p "$CONFIG_DIR"
go run -C "$REPO" ./gno.land/cmd/gnoland config init -config-path "$CONFIG_DIR/config.toml"
# Patch the generated config to bind to 0.0.0.0 (accessible from Docker host)
if command -v sed >/dev/null 2>&1; then
  sed -i.bak 's|tcp://127.0.0.1:26657|tcp://0.0.0.0:26657|' "$CONFIG_DIR/config.toml"
  sed -i.bak 's|tcp://127.0.0.1:26656|tcp://0.0.0.0:26656|' "$CONFIG_DIR/config.toml"
  rm -f "$CONFIG_DIR/config.toml.bak"
fi
echo "  config written to $CONFIG_DIR/config.toml"

# ---- 4. stage genesis.json next to the node data ----
cp "$GENESIS" "$HOME_DIR/genesis.json"

echo ""
echo "done — node home ready at $HOME_DIR"
