#!/bin/sh
# Tests sequential mempool throughput by sending N increment transactions
# one after another without sleep. Verifies that all txs are accepted and
# the final counter value matches the expected count.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=../audit/common.sh
. "$SCRIPT_DIR/../audit/common.sh"

SUFFIX=$(date +%s)
PKGPATH="gno.land/r/${KEY_ADDR}/e2e/counter${SUFFIX}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

TX_COUNT=10

echo "⚡ STARTING SEQUENTIAL STRESS TEST ($TX_COUNT tx)"

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

# --- Increment tx file ---
cat > "$TMPDIR/increment.gno" << EOF
package main

import counter "${PKGPATH}"

func main() { counter.Increment() }
EOF

# --- Sequential stress loop ---
FAILED=0
for i in $(seq 1 $TX_COUNT); do
	echo -n "➡️  Tx #$i: "
	RESULT=$(echo "$PASSWORD" | gnokey maketx run \
		-broadcast -chainid "$CHAINID" -remote "$RPC" \
		-gas-fee 1000000ugnot -gas-wanted 3000000 \
		-insecure-password-stdin -quiet \
		-home "$GNOKEY_HOME" \
		"$KEY" "$TMPDIR/increment.gno" 2>&1)
	if echo "$RESULT" | grep -q "OK!"; then
		echo "✅ Sent"
	else
		echo "❌ Failed"; echo "$RESULT"
		FAILED=$((FAILED + 1))
	fi
done

echo "⏳ Waiting for final commit..."
sleep 5

FINAL=$(gnokey query "vm/qeval" \
	-data "${PKGPATH}.Render(\"\")" \
	-remote "$RPC" 2>&1)

echo "🏁 Final Counter Value (raw): $FINAL"

if echo "$FINAL" | grep -q "\"$TX_COUNT\"" && [ "$FAILED" -eq 0 ]; then
	echo "✅ MEMPOOL STRESS OK — all $TX_COUNT txs committed"
else
	echo "❌ FAILED — $FAILED tx errors, expected counter = $TX_COUNT"
	exit 1
fi
