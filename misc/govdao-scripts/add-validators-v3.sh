#!/usr/bin/env bash
# Batch-add operators to the test13 active valset in one govDAO proposal.
#
# Adds every operator in a SINGLE r/sys/validators/v3 proposal, so the
# 24h valset-update cooldown is consumed once for the whole batch
# instead of once per validator. Each operator must already be
# registered in r/gnops/valopers with KeepRunning=true.
#
# Usage:
#   ./add-validators-v3.sh OPADDR[:POWER] [OPADDR[:POWER] ...]
#
# POWER defaults to 1 when omitted. Max 40 operators per proposal (v3 cap).
#
# Example:
#   ./add-validators-v3.sh g1aaa...:10 g1bbb...:10 g1ccc...
#
# Environment: see README.md.
set -eo pipefail

if [ $# -eq 0 ]; then
  echo "Usage: $0 OPADDR[:POWER] [OPADDR[:POWER] ...]" >&2
  exit 1
fi

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

# v3 rejects more than 40 valoper changes per proposal.
if [ $# -gt 40 ]; then
  echo "ERROR: max 40 operators per proposal (got $#)" >&2
  exit 1
fi

# Parse each "OPADDR[:POWER]" argument into a ValoperChange line for the
# Gno code and a human-readable summary line. POWER defaults to 1.
CHANGES=""
SUMMARY=""
for arg in "$@"; do
  addr="${arg%%:*}"
  power="${arg#*:}"
  [ "$power" = "$arg" ] && power=1

  if [ -z "$addr" ]; then
    echo "ERROR: empty operator address in argument '${arg}'" >&2
    exit 1
  fi

  # Power=0 is a remove in v3; this script only adds, so require a
  # positive integer and point removes at the dedicated script.
  if ! [[ "$power" =~ ^[1-9][0-9]*$ ]]; then
    echo "ERROR: voting power for '${addr}' must be a positive integer (got: '${power}')" >&2
    echo "       use ./rm-validator-v3.sh to remove an operator from the valset" >&2
    exit 1
  fi

  CHANGES="${CHANGES}			valv3.NewValoperChange(address(\"${addr}\"), ${power}),
"
  SUMMARY="${SUMMARY}  ${addr} (power=${power})
"
done

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/add_validators.gno" <<GOEOF
package main

import (
	"gno.land/r/gov/dao"
	valv3 "gno.land/r/sys/validators/v3"
)

func main(cur realm) {
	r := valv3.NewValidatorProposalRequest(
		cross(cur),
		[]valv3.ValoperChange{
${CHANGES}		},
		"Add $# validator(s) to the valset",
		"Batch-add $# operator(s) to the active valset.",
	)
	pid := dao.MustCreateProposal(cross(cur), r)
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	dao.ExecuteProposal(cross(cur), pid)
}
GOEOF

echo "Adding $# validator(s):"
printf '%s' "$SUMMARY"
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
  "$TMPDIR/add_validators.gno"

echo ""
echo "Done — $# validator(s) added."
