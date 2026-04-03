#!/usr/bin/env bash
# Update the description of a registered valoper.
#
# The caller must be the original registrant or on the valoper's auth list.
#
# Usage:
#   ./update-valoper-description.sh <address> <new_description>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <address> <new_description>" >&2
  exit 1
fi

ADDR="$1"
DESCRIPTION="$2"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/update_description.gno" <<GOEOF
package main

import (
	valopers "gno.land/r/gnops/valopers"
)

func main() {
	valopers.UpdateDescription(cross, address("${ADDR}"), "${DESCRIPTION}")
}
GOEOF

echo "Updating description for: ${ADDR}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/update_description.gno"

echo ""
echo "Done — description updated."
