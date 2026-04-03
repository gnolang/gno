#!/usr/bin/env bash
# Extend govDAO T1 membership via MsgRun (requires existing T1 member key).
# Adds 6 new T1 members with 3 invitation points each.
#
# Usage:
#   ./extend-govdao-t1.sh
#
# Environment: see env file. Override inline: VAR=value ./script.sh
set -eo pipefail

# shellcheck source=env
source "$(dirname "$0")/env"

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

func main() {
	ms := memberstore.Get()
	must(ms.SetMember(memberstore.T1, address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"), &memberstore.Member{InvitationPoints: 3})) // Jae
	must(ms.SetMember(memberstore.T1, address("g1m0rgan0rla00ygmdmp55f5m0unvsvknluyg2a4"), &memberstore.Member{InvitationPoints: 3})) // Morgan
	must(ms.SetMember(memberstore.T1, address("g1mx4pum9976th863jgry4sdjzfwu03qan5w2v9j"), &memberstore.Member{InvitationPoints: 3})) // Ray
	must(ms.SetMember(memberstore.T1, address("g12vx7dn3dqq89mz550zwunvg4qw6epq73d9csay"), &memberstore.Member{InvitationPoints: 3})) // Dongwon
	must(ms.SetMember(memberstore.T1, address("g127l4gkhk0emwsx5tmxe96sp86c05h8vg5tufzq"), &memberstore.Member{InvitationPoints: 3})) // Maxwell
	must(ms.SetMember(memberstore.T1, address("g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"), &memberstore.Member{InvitationPoints: 3})) // Milos
}
GOEOF

echo "Extending govDAO T1 with 6 new members"
echo "  Key: ${GNOKEY_NAME}"
echo "  Chain: ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey_run "$TMPDIR/extend_govdao.gno"

echo ""
echo "Done — 6 T1 members added (Jae, Morgan, Ray, Dongwon, Maxwell, Milos)."
