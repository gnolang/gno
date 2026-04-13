#!/usr/bin/env bash
# Update the valoper registration minimum fee via govDAO proposal.
#
# Usage:
#   ./set-valoper-minfee.sh <amount_ugnot>
#   ./set-valoper-minfee.sh 0          # disable registration fee
#   ./set-valoper-minfee.sh 20000000   # set to 20 GNOT
#
# Environment:
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: gnoland1)
#   REMOTE        - RPC endpoint (default: https://rpc.betanet.testnets.gno.land:443)
#   GAS_WANTED    - gas limit (default: 50000000)
#   GAS_FEE       - gas fee (default: 1000000ugnot)
set -eo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <amount_ugnot>" >&2
  echo "       $0 0          # disable registration fee" >&2
  echo "       $0 20000000   # set to 20 GNOT" >&2
  exit 1
fi

MIN_FEE="$1"

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-gnoland1}"
REMOTE="${REMOTE:-https://rpc.betanet.testnets.gno.land:443}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/set_minfee.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/gnops/valopers/proposal"
)

func main() {
	r := proposal.ProposeNewMinFeeProposalRequest(cross, int64(${MIN_FEE}))
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Setting valoper registration min fee to: ${MIN_FEE} ugnot"
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
  "$TMPDIR/set_minfee.gno"

echo ""
echo "Done — valoper min fee updated to ${MIN_FEE} ugnot."
