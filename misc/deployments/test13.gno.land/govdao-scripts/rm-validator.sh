#!/usr/bin/env bash
# Remove a validator from test-13 via govDAO proposal, using
# r/sys/validators/v3. Mirror of add-validator.sh — a voting_power of 0
# signals "remove" in v3's NewProposalRequest.
#
# Usage:
#   ./rm-validator.sh <address>
#
# Environment:
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: test-13)
#   REMOTE        - RPC endpoint (default: http://127.0.0.1:26657)
#   GAS_WANTED    - gas limit (default: 50000000)
#   GAS_FEE       - gas fee (default: 1000000ugnot)
#
# Preconditions:
#   • The target address MUST currently be a validator in v3. v3's PoA
#     backend panics ("proposed validator must be part of the set already")
#     when asked to remove a non-member; the whole proposal aborts and no
#     valset change applies. Use add-validator.sh first if unsure.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-test-13}"
REMOTE="${REMOTE:-http://127.0.0.1:26657}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

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
	valr "gno.land/r/sys/validators/v3"
)

func main() {
	r := valr.NewProposalRequest(
		func() []validators.Validator {
			return []validators.Validator{
				{
					Address:     address("${ADDR}"),
					VotingPower: 0, // 0 = remove
				},
			}
		},
		"Remove validator ${ADDR}",
		"Remove validator ${ADDR} from the valset.",
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

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/rm_validator.gno"
