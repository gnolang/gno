#!/usr/bin/env bash
# Convert a validator bech32 public key (gpub1...) to its gno address (g1...).
#
# Both key types are supported:
#   - secp256k1: regular gnokey secrets-based validator key
#   - ed25519:   gnokms-backed validator key (gnokey validator)
#
# Usage:
#   ./utils/valkey-to-addr.sh <gpub1...>
#
# Example:
#   ./utils/valkey-to-addr.sh gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqdfdtl575xtckdfsjhjwxex2ltwjq7mq36c4y8s4dzcg4gka5pnkq03vsd
#   # => g1u4z9tu4q5838zy07yrd97uu95mkgh4sz5phzsc
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <gpub1...>"
  exit 1
fi

go run -C "$REPO_ROOT" "$SCRIPT_DIR/valkey-to-addr/" "$1"
