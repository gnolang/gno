#!/usr/bin/env bash
# Unrestrict an account so it can transfer ugnot even when bank is locked.
#
# Usage:
#   ./unrestrict-account.sh ADDR [ADDR...]
#
# Example:
#   ./unrestrict-account.sh g1abc...123
#   ./unrestrict-account.sh g1abc...123 g1def...456
#
# Environment:
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: gnoland1)
#   REMOTE        - RPC endpoint (default: https://rpc.betanet.testnets.gno.land:443)
#   GAS_WANTED    - gas limit (default: 50000000)
#   GAS_FEE       - gas fee (default: 1000000ugnot)
set -eo pipefail

if [ $# -eq 0 ]; then
  echo "Usage: $0 ADDR [ADDR...]" >&2
  exit 1
fi

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-gnoland1}"
REMOTE="${REMOTE:-https://rpc.betanet.testnets.gno.land:443}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

# Build address list for the Gno code.
ADDR_ARGS=""
for addr in "$@"; do
  ADDR_ARGS="${ADDR_ARGS}		address(\"${addr}\"),
"
done

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/unrestrict.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/params"
)

func main() {
	r := params.ProposeAddUnrestrictedAcctsRequest(
${ADDR_ARGS}	)
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Unrestricting $# account(s):"
for addr in "$@"; do
  echo "  $addr"
done

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/unrestrict.gno"

echo ""
echo "Done — $# account(s) unrestricted."
