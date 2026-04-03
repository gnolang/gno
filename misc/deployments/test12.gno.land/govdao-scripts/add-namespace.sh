#!/usr/bin/env bash
# Register a custom namespace in r/sys/names/v2 via a govDAO proposal.
#
# Usage:
#   ./add-namespace.sh <namespace> <address>
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=env
source "$(dirname "$0")/env"

if [ $# -ne 2 ]; then
  echo "Usage: $0 <namespace> <address>"
  exit 1
fi

NAMESPACE="$1"
ADDR="$2"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/add_namespace.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	namesv2 "gno.land/r/sys/names/v2"
)

func main() {
	r := namesv2.NewNamespacePropRequest("${NAMESPACE}", address("${ADDR}"))
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Registering namespace: ${NAMESPACE} → ${ADDR}"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/add_namespace.gno"
