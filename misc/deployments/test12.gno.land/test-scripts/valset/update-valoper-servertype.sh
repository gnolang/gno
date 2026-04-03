#!/usr/bin/env bash
# Update the server type of a registered valoper.
#
# The caller must be the original registrant or on the valoper's auth list.
#
# Usage:
#   ./update-valoper-servertype.sh <address> <server_type>
#
# server_type: cloud | on-prem | data-center
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <address> <server_type>" >&2
  echo "server_type: cloud | on-prem | data-center" >&2
  exit 1
fi

ADDR="$1"
SERVER_TYPE="$2"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/update_servertype.gno" <<GOEOF
package main

import (
	valopers "gno.land/r/gnops/valopers"
)

func main() {
	valopers.UpdateServerType(cross, address("${ADDR}"), "${SERVER_TYPE}")
}
GOEOF

echo "Updating server type for: ${ADDR}"
echo "  Server type: ${SERVER_TYPE}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/update_servertype.gno"

echo ""
echo "Done — server type updated to '${SERVER_TYPE}'."
