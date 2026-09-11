#!/usr/bin/env bash
# Unlock ugnot transfers globally so every account can transfer (clears the bank lock).
#
# Usage:
#   ./unlock-transfer.sh
#
# This clears the bank "restricted_denoms" param chain-wide. Once unlocked, the
# per-account unrestrict-account whitelist no longer matters — everyone can
# transfer ugnot. Use lock-transfer (ProposeLockTransferRequest) to re-lock.
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

cat >"$TMPDIR/unlock.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main(cur realm) {
	r := params.ProposeUnlockTransferRequest(cross(cur))
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	dao.ExecuteProposal(cross(cur), pid)
}
GOEOF

echo "Unlocking ugnot transfers globally:"
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
  "$TMPDIR/unlock.gno"

echo ""
echo "Done — ugnot transfers unlocked globally."
