#!/bin/sh
# Targets: fix(gnovm): deep-copy array elements in ArrayValue.Copy (c64feef1d)
# Verifies that local := arr produces an independent copy.
# Without the fix, the local variable aliased the stored array pointer,
# so modifying the copy silently corrupted the original persistent state.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
. "$SCRIPT_DIR/common.sh"

SUFFIX=$(date +%s)
PKGPATH="gno.land/r/${KEY_ADDR}/audit/arrayalias${SUFFIX}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🧪 c64feef1d — Array copy independence"
echo "   Package: $PKGPATH"

cat > "$TMPDIR/arrayalias.gno" << EOF
package arrayalias

import "strconv"

var arr [3]int

func ModifyLocalCopy() {
	local := arr
	local[0] = 999
}

func Render(_ string) string {
	return strconv.Itoa(arr[0])
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

cat > "$TMPDIR/call.gno" << EOF
package main

import a "${PKGPATH}"

func main() { a.ModifyLocalCopy() }
EOF

echo -n "   Calling ModifyLocalCopy()... "
CALL=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot -gas-wanted 5000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/call.gno" 2>&1)
if echo "$CALL" | grep -q "OK!"; then echo "OK"; else
	echo "FAILED"; echo "$CALL"; exit 1
fi

echo -n "   Querying arr[0] (expect 0)... "
RESULT=$(gnokey query "vm/qeval" \
	-data "${PKGPATH}.Render(\"\")" \
	-remote "$RPC" 2>&1)

if echo "$RESULT" | grep -q '"0"'; then
	echo "✅ PATCHED — arr[0] unchanged after copy modification"
elif echo "$RESULT" | grep -q '"999"'; then
	echo "❌ VULNERABLE — arr[0] aliased and corrupted to 999"
	exit 1
else
	echo "⚠️  UNKNOWN OUTPUT"; echo "$RESULT"; exit 1
fi
