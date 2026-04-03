#!/usr/bin/env bash
# Register a valoper in the r/gnops/valopers registry.
#
# Usage:
#   ./register-valoper.sh <pub_key> <moniker> <description> <server_type>
#
# server_type: cloud | on-prem | data-center
#
# The validator address is derived from the public key automatically.
# The registration fee (VALOPER_REGISTRATION_FEE, default 20 GNOT) is paid
# from GNOKEY_NAME's balance.
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -lt 4 ]; then
  echo "Usage: $0 <pub_key> <moniker> <description> <server_type>" >&2
  echo "" >&2
  echo "server_type: cloud | on-prem | data-center" >&2
  echo "" >&2
  echo "Example:" >&2
  echo "  $0 gpub1pgfj7... 'MyNode' 'A reliable validator' cloud" >&2
  exit 1
fi

PUB_KEY="$1"
MONIKER="$2"
DESCRIPTION="$3"
SERVER_TYPE="$4"

ADDR=$(pubkey_to_addr "$PUB_KEY")

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/register_valoper.gno" <<GOEOF
package main

import (
	valopers "gno.land/r/gnops/valopers"
)

func main() {
	valopers.Register(cross, "${MONIKER}", "${DESCRIPTION}", "${SERVER_TYPE}", address("${ADDR}"), "${PUB_KEY}")
}
GOEOF

echo "Registering valoper: ${ADDR} (moniker=${MONIKER})"
echo "  PubKey: ${PUB_KEY}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo "  Fee: ${VALOPER_REGISTRATION_FEE} ugnot"
echo ""

gnokey_run "$TMPDIR/register_valoper.gno" -send "${VALOPER_REGISTRATION_FEE}ugnot"

echo ""
echo "Done — valoper ${ADDR} registered as '${MONIKER}'."
