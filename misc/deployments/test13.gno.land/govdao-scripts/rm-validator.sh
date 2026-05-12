#!/usr/bin/env bash
# Remove an operator from the test-13 active valset via govDAO proposal.
#
# Routes through r/sys/validators/v3's operator-keyed
# NewValidatorProposalRequest (non-crossing) with Power=0. This is a
# force-remove — unlike the higher-level facade in
# r/gnops/valopers/proposal.NewValidatorProposalRequest, this does
# not require the operator to have called UpdateKeepRunning(false)
# first. The operator must still exist in r/gnops/valopers'
# valoperCache (v3 enforces that at proposal-creation time).
#
# Usage:
#   ./rm-validator.sh <operator_address>
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:-aeddi}"
CHAIN_ID="${CHAIN_ID:-test-13}"
REMOTE="${REMOTE:-127.0.0.1:26657}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <operator_address>"
  echo ""
  echo "Example:"
  echo "  $0 g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr"
  exit 1
fi

ADDR="$1"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/rm_validator.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	valv3 "gno.land/r/sys/validators/v3"
)

func main() {
	r := valv3.NewValidatorProposalRequest(
		[]valv3.ValoperChange{
			{OperatorAddress: address("${ADDR}"), Power: 0},
		},
		"Remove validator ${ADDR}",
		"Remove operator ${ADDR} from the active valset.",
	)
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Removing validator: ${ADDR}"
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
  "$TMPDIR/rm_validator.gno"
