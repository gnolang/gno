#!/usr/bin/env bash
# Disable the valset-update cooldown entirely via govDAO proposal.
#
# Sets r/sys/validators/v3's valset-update cooldown to 0 seconds (the default
# is 24h), removing the minimum interval between consecutive validator-set
# updates. Useful on testnets where validators are added/removed rapidly.
# Re-enable later by proposing a non-zero cooldown via NewCooldownPropRequest.
#
# Usage:
#   ./disable-valset-cooldown.sh
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/disable_cooldown.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	valv3 "gno.land/r/sys/validators/v3"
)

func main(cur realm) {
	r := valv3.NewCooldownPropRequest(
		cross(cur),
		0,
		"Disable valset update cooldown",
		"Set the valset-update cooldown to 0 seconds, removing the 24h minimum interval between consecutive validator-set updates.",
	)
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	dao.ExecuteProposal(cross(cur), pid)
}
GOEOF

echo "Disabling valset-update cooldown (setting to 0s) via govDAO proposal:"
echo "  Key:    ${GNOKEY_NAME}"
echo "  Chain:  ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/disable_cooldown.gno"

echo ""
echo "Done — valset-update cooldown disabled (0s)."
