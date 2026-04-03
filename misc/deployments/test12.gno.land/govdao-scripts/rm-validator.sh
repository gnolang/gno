#!/usr/bin/env bash
# Remove a validator from test12 via govDAO proposal.
#
# Usage:
#   ./rm-validator.sh <address>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=common
source "$(dirname "$0")/common"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <address>"
  echo ""
  echo "Example:"
  echo "  $0 g1abc...xyz"
  exit 1
fi

ADDR="$1"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/rm_validator.gno" <<GOEOF
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
					VotingPower: 0,
				},
			}
		},
		"Remove validator ${ADDR}",
		"Remove validator ${ADDR} from the validator set",
	)
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Removing validator: ${ADDR}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/rm_validator.gno"
