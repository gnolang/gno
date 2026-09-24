#!/bin/sh
# Tests cross-validator state consistency via a simple counter realm.
# Deploys a fresh counter realm, sends an Increment tx, then queries the
# node to verify the state was committed correctly.
# Note: in a multi-validator setup, query both nodes and assert equal state.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=../audit/common.sh
. "$SCRIPT_DIR/../audit/common.sh"

SUFFIX=$(date +%s)
PKGPATH="gno.land/r/${KEY_ADDR}/e2e/counter${SUFFIX}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🚀 E2E COUNTER TEST"

# --- Deploy counter realm ---
cat > "$TMPDIR/counter.gno" << EOF
package counter

import "strconv"

type state struct{ count int }

func (s *state) inc() { s.count++ }

var counter state

func Increment() { counter.inc() }

func Render(_ string) string {
	return strconv.Itoa(counter.count)
}
EOF

cat > "$TMPDIR/gnomod.toml" << EOF
module = "${PKGPATH}"
gno = "0.9"
EOF

echo -n "   Deploying counter realm... "
DEPLOY=$(echo "$PASSWORD" | gnokey maketx addpkg \
	-pkgpath "$PKGPATH" -pkgdir "$TMPDIR" \
	-gas-fee 1000000ugnot -gas-wanted 10000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" 2>&1)
if echo "$DEPLOY" | grep -q "OK!"; then echo "OK"; else
	echo "FAILED"; echo "$DEPLOY"; exit 1
fi

# --- Increment tx ---
cat > "$TMPDIR/increment.gno" << EOF
package main

import counter "${PKGPATH}"

func main() { counter.Increment() }
EOF

echo -n "➡️  Sending Increment tx... "
INC=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot -gas-wanted 3000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/increment.gno" 2>&1)
if echo "$INC" | grep -q "OK!"; then echo "OK"; else
	echo "FAILED"; echo "$INC"; exit 1
fi

sleep 2

# --- Query and verify ---
echo -n "🔍 Querying counter state (expect 1)... "
RESULT=$(gnokey query "vm/qeval" \
	-data "${PKGPATH}.Render(\"\")" \
	-remote "$RPC" 2>&1)

if echo "$RESULT" | grep -q '"1"'; then
	echo "✅ E2E COUNTER OK — state = 1"
else
	echo "❌ FAILED — unexpected state"; echo "$RESULT"; exit 1
fi
