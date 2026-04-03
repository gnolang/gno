#!/usr/bin/env bash
# Add a validator to test12 via govDAO proposal.
#
# Usage:
#   ./add-validator.sh <pub_key> [voting_power]
#
# The validator address is derived from the public key automatically.
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <pub_key> [voting_power]"
  echo ""
  echo "Example:"
  echo "  $0 gpub1pggj7... 1"
  exit 1
fi

PUB_KEY="$1"
POWER="${2:-1}"

ADDR=$(pubkey_to_addr "$PUB_KEY")

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
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Adding validator: ${ADDR} (power=${POWER})"
echo "  PubKey: ${PUB_KEY}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/add_validator.gno"
