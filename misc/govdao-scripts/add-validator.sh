#!/usr/bin/env bash
# Add a validator via govDAO proposal.
#
# Usage:
#   ./add-validator.sh <address> <pub_key> [voting_power]
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

if [ $# -lt 2 ]; then
  echo "Usage: $0 <address> <pub_key> [voting_power]"
  echo ""
  echo "Example:"
  echo "  $0 g1abc...xyz gpub1pggj7... 1"
  exit 1
fi

ADDR="$1"
PUB_KEY="$2"
POWER="${3:-1}"

# voting_power must be a positive integer. Tendermint treats Power=0 as
# a remove operation, which would silently turn this script into a
# remove if a user passes 0 by mistake. Catch it here.
if ! [[ "$POWER" =~ ^[1-9][0-9]*$ ]]; then
  echo "ERROR: voting_power must be a positive integer (got: '$POWER')"
  echo "       use ./rm-validator.sh to remove a validator from the set"
  exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/add_validator.gno" <<GOEOF
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	valr "gno.land/r/sys/validators/v2"
)

func main() {
	r := valr.NewPropRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("${ADDR}"),
					PubKey:      "${PUB_KEY}",
					VotingPower: ${POWER},
				},
			}
		},
		"Add validator ${ADDR}",
		"Add validator ${ADDR} with power ${POWER}",
	)
	pid := dao.MustCreateProposal(cross1, r)
	dao.MustVoteOnProposal(cross1, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross1, pid)
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
