#!/usr/bin/env bash
# Update the moniker of a registered valoper.
#
# The caller must be the original registrant or on the valoper's auth list.
#
# Usage:
#   ./update-valoper-moniker.sh <address> <new_moniker>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <address> <new_moniker>" >&2
  exit 1
fi

ADDR="$1"
MONIKER="$2"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/update_moniker.gno" <<GOEOF
package main

import (
	valopers "gno.land/r/gnops/valopers"
)

func main() {
	valopers.UpdateMoniker(cross, address("${ADDR}"), "${MONIKER}")
}
GOEOF

echo "Updating moniker for: ${ADDR}"
echo "  New moniker: ${MONIKER}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/update_moniker.gno"

echo ""
echo "Done — moniker updated to '${MONIKER}'."
