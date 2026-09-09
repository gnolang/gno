#!/bin/sh
# Targets: fix(gnovm): call DidUpdate on DataByte index assignment (NEWTENDG-98, a3a356e71)
# Verifies that bs[i] = v mutations persist across transactions.
# Without the fix, byte-slice index writes were silently dropped after tx commit.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
. "$SCRIPT_DIR/common.sh"

SUFFIX=$(date +%s)
PKGPATH="gno.land/r/${KEY_ADDR}/audit/byteslice${SUFFIX}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🧪 NEWTENDG-98 — Byte-slice index mutation persistence"
echo "   Package: $PKGPATH"

cat > "$TMPDIR/byteslice.gno" << EOF
package byteslice

import "strconv"

type ByteState struct {
	data []byte
}

func (b *ByteState) set(i int, v byte) {
	b.data[i] = v
}

var state = ByteState{data: []byte{0, 0, 0}}

func SetFirst(v byte) {
	state.set(0, v)
}

func Render(_ string) string {
	return strconv.Itoa(int(state.data[0]))
}
EOF

cat > "$TMPDIR/gnomod.toml" << EOF
module = "${PKGPATH}"
gno = "0.9"
EOF

echo -n "   Deploying realm... "
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

cat > "$TMPDIR/set.gno" << EOF
package main

import byteslice "${PKGPATH}"

func main() { byteslice.SetFirst(5) }
EOF

echo -n "   Setting bs[0] = 5... "
SET=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot -gas-wanted 5000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/set.gno" 2>&1)
if echo "$SET" | grep -q "OK!"; then echo "OK"; else
	echo "FAILED"; echo "$SET"; exit 1
fi

echo -n "   Querying bs[0] (expect 5)... "
RESULT=$(gnokey query "vm/qeval" \
	-data "${PKGPATH}.Render(\"\")" \
	-remote "$RPC" 2>&1)

if echo "$RESULT" | grep -q '"5"'; then
	echo "✅ PATCHED — bs[0] = 5 persisted correctly"
elif echo "$RESULT" | grep -q '"0"'; then
	echo "❌ VULNERABLE — bs[0] mutation was silently dropped (still 0)"
	exit 1
else
	echo "⚠️  UNKNOWN OUTPUT"; echo "$RESULT"; exit 1
fi
