#!/usr/bin/env bash
# Set or update the CLA document via govDAO proposal.
# Downloads the URL, computes sha256, and submits the proposal.
#
# Usage:
#   ./set-cla.sh URL
#   ./set-cla.sh ""   # disable CLA enforcement
#
# Example:
#   ./set-cla.sh https://raw.githubusercontent.com/gnolang/gno/.../CLA.md
#
# Environment:
#   GNOKEY_NAME   - gnokey key name (default: moul)
#   CHAIN_ID      - chain ID (default: gnoland1)
#   REMOTE        - RPC endpoint (default: https://rpc.betanet.testnets.gno.land:443)
#   GAS_WANTED    - gas limit (default: 50000000)
#   GAS_FEE       - gas fee (default: 1000000ugnot)
set -eo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 URL" >&2
  echo "       $0 \"\"    # disable CLA enforcement" >&2
  exit 1
fi

CLA_URL="$1"

GNOKEY_NAME="${GNOKEY_NAME:-moul}"
CHAIN_ID="${CHAIN_ID:-gnoland1}"
REMOTE="${REMOTE:-https://rpc.betanet.testnets.gno.land:443}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

sha256_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | cut -d' ' -f1
  else
    echo "Error: no sha256 tool found (install coreutils or perl)" >&2
    return 1
  fi
}

if [ -z "$CLA_URL" ]; then
  CLA_HASH=""
  echo "Disabling CLA enforcement"
else
  echo "Fetching CLA from: $CLA_URL"
  wget -q -O "$TMPDIR/cla.md" "$CLA_URL"
  CLA_HASH=$(sha256_file "$TMPDIR/cla.md")
  echo "  sha256: $CLA_HASH"
fi

cat >"$TMPDIR/set_cla.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/cla"
)

func main() {
	r := cla.ProposeNewCLA("${CLA_HASH}", "${CLA_URL}")
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

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
  "$TMPDIR/set_cla.gno"

echo ""
echo "Done — CLA updated."
