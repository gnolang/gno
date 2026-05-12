#!/usr/bin/env bash
# Add one or more addresses to the unrestricted-accounts set so they
# can transfer ugnot even while the bank is in restricted-denom mode
# (the regime test-13 enters at genesis via govdao_prop1_test13's
# ProposeLockTransferRequest).
#
# Routes through r/sys/params.ProposeAddUnrestrictedAcctsRequest,
# the same proposal used by phase-1 bootstrap to whitelist the
# GovDAO multisig + 10 test-13 faucets.
#
# Usage:
#   ./unrestrict-account.sh <address> [<address>…]
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
CHAIN_ID="${CHAIN_ID:-test-13}"
REMOTE="${REMOTE:-127.0.0.1:26657}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <address> [<address>…]"
  echo ""
  echo "Example:"
  echo "  $0 g1abc…xyz                       # single address"
  echo "  $0 g1abc…xyz g1def…uvw             # two at once"
  exit 1
fi

# Build the variadic address list for the .gno body.
ADDR_ARGS=""
for addr in "$@"; do
  ADDR_ARGS="${ADDR_ARGS}		address(\"${addr}\"),
"
done

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/unrestrict_account.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main() {
	r := params.ProposeAddUnrestrictedAcctsRequest(
${ADDR_ARGS}	)
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Unrestricting $# account(s):"
for addr in "$@"; do
  echo "  ${addr}"
done
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
  "$TMPDIR/unrestrict_account.gno"
