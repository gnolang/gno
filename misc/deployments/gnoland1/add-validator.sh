#!/usr/bin/env bash
# Add a validator to gnoland1 via govDAO proposal.
#
# Usage:
#   ./add-validator.sh <address> <pub_key> [voting_power]
#
# Environment:
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: gnoland1)
#   REMOTE        - RPC endpoint (default: 127.0.0.1:26657)
#   GAS_WANTED    - gas limit (default: 10000000)
#   GAS_FEE       - gas fee (default: 1000000ugnot)
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-gnoland1}"
#REMOTE="${REMOTE:-https://rpc.betanet.testnets.gno.land:443}"
REMOTE="${REMOTE:-https://sentry1.gnoland1.gno.berty.io:26657}"
GAS_WANTED="${GAS_WANTED:-100000000}"
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
  "$TMPDIR/add_validator.gno"
