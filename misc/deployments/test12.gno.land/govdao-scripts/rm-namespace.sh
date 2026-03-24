#!/usr/bin/env bash
# Remove a custom namespace from r/sys/names/v2 via a govDAO proposal.
#
# Usage:
#   ./rm-namespace.sh <namespace>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=env
source "$(dirname "$0")/env"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <namespace>"
  exit 1
fi

NAMESPACE="$1"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/rm_namespace.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	namesv2 "gno.land/r/sys/names/v2"
)

func main() {
	r := namesv2.NewNamespaceRemovalPropRequest("${NAMESPACE}")
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Removing namespace: ${NAMESPACE}"
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
  "$TMPDIR/rm_namespace.gno"
