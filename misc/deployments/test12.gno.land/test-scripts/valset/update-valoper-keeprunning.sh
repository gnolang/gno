#!/usr/bin/env bash
# Update the KeepRunning flag of a registered valoper.
#
# KeepRunning=true signals intent to stay in the validator set.
# KeepRunning=false signals intent to leave; a subsequent govDAO proposal
# via add-validator-from-valopers.sh will remove the validator from the set.
#
# The caller must be the original registrant or on the valoper's auth list.
#
# Usage:
#   ./update-valoper-keeprunning.sh <address> <true|false>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <address> <true|false>" >&2
  exit 1
fi

ADDR="$1"
KEEP_RUNNING="$2"

if [ "$KEEP_RUNNING" != "true" ] && [ "$KEEP_RUNNING" != "false" ]; then
  echo "Error: keep_running must be 'true' or 'false'" >&2
  exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/update_keeprunning.gno" <<GOEOF
package main

import (
	valopers "gno.land/r/gnops/valopers"
)

func main() {
	valopers.UpdateKeepRunning(cross, address("${ADDR}"), ${KEEP_RUNNING})
}
GOEOF

echo "Updating KeepRunning for: ${ADDR}"
echo "  KeepRunning: ${KEEP_RUNNING}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/update_keeprunning.gno"

echo ""
echo "Done — KeepRunning set to ${KEEP_RUNNING}."
