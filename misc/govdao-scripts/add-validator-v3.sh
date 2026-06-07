#!/usr/bin/env bash
# Add an operator to the test-13 active valset via govDAO proposal.
#
# Routes through r/sys/validators/v3's operator-keyed
# NewValidatorProposalRequest (non-crossing) with the given Power.
# The operator must already exist in r/gnops/valopers' valoperCache
# (i.e. have called valopers.Register themselves) with
# KeepRunning=true (the default at Register time). v3's executor
# re-resolves the signing pubkey from the cache at execution time,
# so a mid-flight key rotation publishes the current key.
#
# Usage:
#   ./add-validator.sh <operator_address> [voting_power]
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <operator_address> [voting_power]"
  echo ""
  echo "Example:"
  echo "  $0 g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr        # power 1 (default)"
  echo "  $0 g1s2ht24e85qq3t66gc9sgdvk5kzc38yy68aaqvr 10     # power 10"
  exit 1
fi

ADDR="$1"
POWER="${2:-1}"

# voting_power must be a positive integer — v3 rejects Power=0 as a
# remove operation, which would silently turn this script into a
# remove if a user passes 0 by mistake. Catch it here.
if ! [[ "$POWER" =~ ^[1-9][0-9]*$ ]]; then
  echo "ERROR: voting_power must be a positive integer (got: '$POWER')"
  echo "       use ./rm-validator.sh to remove an operator from the valset"
  exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/add_validator.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	valv3 "gno.land/r/sys/validators/v3"
)

func main(cur realm) {
	r := valv3.NewValidatorProposalRequest(
		[]valv3.ValoperChange{
			{OperatorAddress: address("${ADDR}"), Power: ${POWER}},
		},
		"Add validator ${ADDR}",
		"Add operator ${ADDR} to the active valset with voting power ${POWER}.",
	)
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposal(cross(cur), dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross(cur), pid)
}
GOEOF

echo "Adding validator: ${ADDR} (power=${POWER})"
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
  "$TMPDIR/add_validator.gno"
