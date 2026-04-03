#!/usr/bin/env bash
# Add a member to a valoper's auth list.
#
# Auth list members can call Update* functions on behalf of the valoper.
# Only the original registrant (owner) can modify the auth list.
#
# Usage:
#   ./add-auth-member.sh <valoper_address> <member_address>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <valoper_address> <member_address>" >&2
  exit 1
fi

VALOPER_ADDR="$1"
MEMBER_ADDR="$2"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/add_auth_member.gno" <<GOEOF
package main

import (
	valopers "gno.land/r/gnops/valopers"
)

func main() {
	valopers.AddToAuthList(cross, address("${VALOPER_ADDR}"), address("${MEMBER_ADDR}"))
}
GOEOF

echo "Adding auth member to valoper: ${VALOPER_ADDR}"
echo "  Member: ${MEMBER_ADDR}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/add_auth_member.gno"

echo ""
echo "Done — ${MEMBER_ADDR} added to auth list of ${VALOPER_ADDR}."
