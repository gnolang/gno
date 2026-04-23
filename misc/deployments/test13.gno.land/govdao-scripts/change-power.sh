#!/usr/bin/env bash
# Update the voting power of an existing test-13 validator.
#
# v3's PoA backend doesn't expose an in-place update — AddValidator returns
# "validator already in set" when called for a known address. This script
# works around that by batching remove + re-add in a single govDAO proposal
# so the change appears to consensus as one valset diff (the combined
# EndBlocker diff sees the power transition atomically, not as a hole
# between two blocks).
#
# Usage:
#   ./change-power.sh <address> <pub_key> <new_voting_power>
#
# Environment (same defaults as add-validator.sh):
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: test-13)
#   REMOTE        - RPC endpoint (default: http://127.0.0.1:26657)
#   GAS_WANTED    - gas limit (default: 50000000)
#   GAS_FEE       - gas fee (default: 1000000ugnot)
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-test-13}"
REMOTE="${REMOTE:-http://127.0.0.1:26657}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

if [ $# -lt 3 ]; then
  echo "Usage: $0 <address> <pub_key> <new_voting_power>"
  echo ""
  echo "Example:"
  echo "  $0 g1abc...xyz gpub1pggj7... 5"
  exit 1
fi

ADDR="$1"
PUB_KEY="$2"
POWER="$3"

if ! [[ "$POWER" =~ ^[1-9][0-9]*$ ]]; then
  echo "error: new_voting_power must be a positive integer (use rm-validator.sh for removal)" >&2
  exit 1
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/change_power.gno" <<GOEOF
package main

import (
	"gno.land/p/sys/validators"
	"gno.land/r/gov/dao"
	valr "gno.land/r/sys/validators/v3"
)

// Remove the existing entry and re-add at the new power in the same
// executor. EndBlocker computes a single UpdatesFrom diff covering both,
// so the tm2 valset transitions directly to the new power without an
// intermediate "validator absent" step.
func main() {
	executor := valr.NewValsetChangeExecutor(func() []validators.Validator {
		return []validators.Validator{
			{
				Address:     address("${ADDR}"),
				VotingPower: 0, // remove first
			},
			{
				Address:     address("${ADDR}"),
				PubKey:      "${PUB_KEY}",
				VotingPower: ${POWER}, // re-add at new power
			},
		}
	})

	r := dao.NewProposalRequest(
		"Update validator ${ADDR} power → ${POWER}",
		"Atomic remove+add to change ${ADDR}'s voting power.",
		executor,
	)

	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Updating validator: ${ADDR} → power=${POWER}"
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
  "$TMPDIR/change_power.gno"
