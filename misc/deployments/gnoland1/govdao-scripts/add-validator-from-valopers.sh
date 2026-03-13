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
# Environment:
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: gnoland1)
#   REMOTE        - RPC endpoint (default: https://rpc.betanet.testnets.gno.land:443)
#   GAS_WANTED    - gas limit (default: 10000000)
#   GAS_FEE       - gas fee (default: 1000000ugnot)
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-gnoland1}"
REMOTE="${REMOTE:-https://rpc.betanet.testnets.gno.land:443}"
GAS_WANTED="${GAS_WANTED:-10000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

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
