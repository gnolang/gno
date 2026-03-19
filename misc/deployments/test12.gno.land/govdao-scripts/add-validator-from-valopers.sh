#!/usr/bin/env bash
# Add a validator from the r/gnops/valopers registry via govDAO proposal.
#
# Uses r/gnops/valopers/proposal.NewValidatorProposalRequest to look up the
# valoper profile on-chain and create a governance proposal, then votes YES
# and executes it immediately.
#
# Usage:
#   ./add-validator-from-valopers.sh <address>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=env
source "$(dirname "$0")/env"

if [ $# -lt 1 ]; then
  echo "Usage: $0 <address>"
  echo ""
  echo "Looks up the valoper profile from r/gnops/valopers and creates a"
  echo "govDAO proposal to add them to the validator set, votes YES, and"
  echo "executes it."
  echo ""
  echo "The valoper must have registered at r/gnops/valopers first."
  echo ""
  echo "Example:"
  echo "  $0 g1abc...xyz"
  exit 1
fi

ADDR="$1"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/add_from_valopers.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gnops/valopers/proposal"
)

func main() {
	r := proposal.NewValidatorProposalRequest(cross, address("${ADDR}"))
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Adding validator from valopers: ${ADDR}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/add_from_valopers.gno"
