#!/usr/bin/env bash
# Extend govDAO T1 membership via MsgRun (requires existing T1 member key).
# Seats the full T1 roster with 3 invitation points each. Addresses that are
# already members -- including the signer, who must be T1 to authorize this --
# are skipped, so the script is re-runnable and works for whichever T1 member
# runs it (moul on gnoland1/test-13, aeddi on pearl/sapphire/topaz).
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

type rosterEntry struct {
	name string
	addr address
}

// t1Roster is the full target T1 membership, deliberately signer-agnostic:
// the signer is necessarily already a T1 member (that is what authorizes this
// MsgRun), so their own entry is filtered out at runtime instead of being
// hardcoded out of the list.
var t1Roster = []rosterEntry{
	{"Jae", "g1ecsuj0q572jr0dhu29q9njtnmw03hyu7tyyvv6"},
	{"Morgan", "g1m0rgan0rla00ygmdmp55f5m0unvsvknluyg2a4"},
	{"Aeddi", "g1aeddlftlfk27ret5rf750d7w5dume3kcsm8r8m"},
	{"Dongwon", "g1gzhj234kpajz963z5vf42j4ylddscnkez2wvly"},
	{"Maxwell", "g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq"},
	{"Milos", "g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"},
	{"Manfred", "g1manfred47kzduec920z88wfr64ylksmdcedlf5"},
}

func main(cur realm) {
	ms := memberstore.Get(0, cur)
	for _, r := range t1Roster {
		// SetMember errors out if the address already sits in any tier, and a
		// single error would abort the whole transaction. Skip instead: this is
		// what drops the signer's own entry, and it makes reruns idempotent.
		if _, tier := ms.GetMember(r.addr); tier != "" {
			println("skip " + r.name + " -- already " + tier)
			continue
		}
		if err := ms.SetMember(memberstore.T1, r.addr, &memberstore.Member{InvitationPoints: 3}); err != nil {
			panic(err.Error())
		}
		println("seat " + r.name + " as T1")
	}
}
GOEOF

echo "Extending govDAO T1 to the full roster (already-seated addresses, incl. the signer, are skipped)"
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
echo "Done -- see the run output above for which members were seated vs. skipped."
