#!/usr/bin/env bash
# Register a custom username for an address via govDAO proposal.
#
# Routes through r/sys/users.ProposeRegisterUser. The proposal
# bypasses the controller whitelist (where r/sys/namereg/v1 lives)
# and the canonical-collision check (decision #3: DAO grants are
# sovereign). A canonical-collision warning is auto-injected into
# the proposal description if relevant; voters can NO it.
#
# After execution, <address> can deploy packages under
# gno.land/r/<username>/* and gno.land/p/<username>/* — the
# r/sys/names verifier bridges to r/sys/users.ResolveName for
# registered-name namespaces.
#
# Names must match the validateName regex:
#   ^[a-z][a-z0-9]*([_-][a-z0-9]+)*$
#
# Usage:
#   ./register-user.sh <username> <address>
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

if [ $# -lt 2 ]; then
  echo "Usage: $0 <username> <address>"
  echo ""
  echo "Example:"
  echo "  $0 moul g1moul0123456789abcdefghijklmnopqrstuvwxy"
  exit 1
fi

USERNAME="$1"
ADDR="$2"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/register_user.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	"gno.land/r/sys/users"
)

func main() {
	r := users.ProposeRegisterUser("${USERNAME}", address("${ADDR}"))
	pid := dao.MustCreateProposal(cross, r)
	dao.MustVoteOnProposal(cross, dao.VoteRequest{Option: dao.YesVote, ProposalID: pid})
	dao.ExecuteProposal(cross, pid)
}
GOEOF

echo "Registering user: ${USERNAME} -> ${ADDR}"
echo "  Key:    ${GNOKEY_NAME}"
echo "  Chain:  ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/register_user.gno"
