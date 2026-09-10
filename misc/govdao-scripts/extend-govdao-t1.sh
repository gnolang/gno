#!/usr/bin/env bash
# Extend govDAO T1 membership via MsgRun (requires existing T1 member key).
# Adds 6 new T1 members with 3 invitation points each.
#
# Usage:
#   ./extend-govdao-t1.sh
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

cat >"$TMPDIR/extend_govdao.gno" <<'GOEOF'
package main

import (
	"gno.land/r/gov/dao/v3/memberstore"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main(cur realm) {
	ms := memberstore.Get(0, cur)
	must(ms.SetMember(memberstore.T1, address("g1ecsuj0q572jr0dhu29q9njtnmw03hyu7tyyvv6"), &memberstore.Member{InvitationPoints: 3})) // Jae
	must(ms.SetMember(memberstore.T1, address("g1m0rgan0rla00ygmdmp55f5m0unvsvknluyg2a4"), &memberstore.Member{InvitationPoints: 3})) // Morgan
	must(ms.SetMember(memberstore.T1, address("g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m"), &memberstore.Member{InvitationPoints: 3})) // Aeddi
	must(ms.SetMember(memberstore.T1, address("g1gzhj234kpajz963z5vf42j4ylddscnkez2wvly"), &memberstore.Member{InvitationPoints: 3})) // Dongwon
	must(ms.SetMember(memberstore.T1, address("g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq"), &memberstore.Member{InvitationPoints: 3})) // Maxwell
	must(ms.SetMember(memberstore.T1, address("g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"), &memberstore.Member{InvitationPoints: 3})) // Milos
}
GOEOF

echo "Extending govDAO T1 with 6 new members"
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
  "$TMPDIR/extend_govdao.gno"

echo ""
echo "Done — 6 T1 members added (Jae, Morgan, Aeddi, Dongwon, Maxwell, Milos)."
