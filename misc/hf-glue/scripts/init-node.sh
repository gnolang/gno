#!/usr/bin/env bash
# Initialise gnoland-home for the testbed:
#   - run `gnoland secrets init` to generate the single validator identity
#     (skipped in keyless mode — caller already has a remote signer)
#   - rewrite validators in out/genesis.json so it contains ONLY that key
#
# The docker container mounts $OUT/gnoland-home/ as its --data-dir, so the
# node boots with the key generated here.
#
# Two modes:
#
#   priv-key (default) — locally-signed validator. Generates a fresh
#       priv_validator_key.json under $OUT/gnoland-home/secrets/ if absent;
#       the resulting genesis has that key as the sole validator.
#
#   keyless — remote-signed validator (e.g. gnokms). Caller supplies the
#       bech32 address + pubkey of the remote signer's key via env. No
#       priv key is generated. Triggered by setting both
#       VALIDATOR_ADDRESS and VALIDATOR_PUBKEY.
#
# Inputs (env):
#   VALIDATOR_NAME       name baked into the genesis validator entry
#   OUT                  output directory (absolute)
#   REPO                 repo root (absolute)
#   VALIDATOR_ADDRESS    (keyless) bech32 g1... — must derive from PUBKEY
#   VALIDATOR_PUBKEY     (keyless) bech32 gpub1...
set -euo pipefail

: "${VALIDATOR_NAME:?VALIDATOR_NAME is required}"
: "${OUT:?OUT is required}"
: "${REPO:?REPO is required}"

GENESIS="$OUT/genesis.json"
HOME_DIR="$OUT/gnoland-home"
SECRETS_DIR="$HOME_DIR/secrets"
PV_KEY="$SECRETS_DIR/priv_validator_key.json"

VALIDATOR_ADDRESS="${VALIDATOR_ADDRESS:-}"
VALIDATOR_PUBKEY="${VALIDATOR_PUBKEY:-}"
KEYLESS=0
if [[ -n "$VALIDATOR_ADDRESS" || -n "$VALIDATOR_PUBKEY" ]]; then
  if [[ -z "$VALIDATOR_ADDRESS" || -z "$VALIDATOR_PUBKEY" ]]; then
    echo "ERROR: keyless mode requires both VALIDATOR_ADDRESS and VALIDATOR_PUBKEY" >&2
    exit 1
  fi
  KEYLESS=1
fi

if [[ ! -f "$GENESIS" ]]; then
  echo "missing $GENESIS — run 'make fetch' first" >&2
  exit 1
fi

echo "── init single-validator node ───────────────────────────────"
mkdir -p "$HOME_DIR"

# ---- 1. generate / acknowledge validator identity ----
if [[ "$KEYLESS" -eq 1 ]]; then
  echo "  keyless mode (remote signer assumed)"
  echo "    address: $VALIDATOR_ADDRESS"
  echo "    pubkey:  $VALIDATOR_PUBKEY"
elif [[ -f "$PV_KEY" ]]; then
  echo "  secrets already present at $SECRETS_DIR — reusing"
else
  echo "  generating secrets in $SECRETS_DIR"
  mkdir -p "$SECRETS_DIR"
  go run -C "$REPO" ./gno.land/cmd/gnoland secrets init --data-dir "$SECRETS_DIR"
fi

# ---- 2. rewrite validator set in the genesis to a single entry ----
echo ""
echo "  rewriting validator set in genesis..."
if [[ "$KEYLESS" -eq 1 ]]; then
  go run -C "$REPO/misc/hf-glue/fixvalidator" . \
    --address "$VALIDATOR_ADDRESS" \
    --pubkey "$VALIDATOR_PUBKEY" \
    --genesis "$GENESIS" \
    --name "$VALIDATOR_NAME" \
    --power 10
else
  go run -C "$REPO/misc/hf-glue/fixvalidator" . \
    --priv-key "$PV_KEY" \
    --genesis "$GENESIS" \
    --name "$VALIDATOR_NAME" \
    --power 10
fi

# ---- 3. write config.toml so RPC binds to 0.0.0.0 (accessible from host) ----
CONFIG_DIR="$HOME_DIR/config"
mkdir -p "$CONFIG_DIR"
# `gnoland config init` refuses to overwrite an existing file; pass -force so
# re-running `make init` after `make migrate` (the documented workflow) just
# regenerates the config rather than aborting.
go run -C "$REPO" ./gno.land/cmd/gnoland config init -force -config-path "$CONFIG_DIR/config.toml"
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
