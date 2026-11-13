#!/bin/sh
# Targets: fix(gnovm): implement Go-compliant variable initialization order (NEWTENDG-68, 50ee56e64)
# Verifies that package-level vars are initialized in dependency order, not declaration order.
# In Go: var B = A + 1; var A = 2 → A is initialized first → B = 3.
# Without the fix, Gno initialized in declaration order → A = 0 when B was set → B = 1.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
. "$SCRIPT_DIR/common.sh"

SUFFIX=$(date +%s)
PKGPATH="gno.land/r/${KEY_ADDR}/audit/varinit${SUFFIX}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🧪 NEWTENDG-68 — Package-level variable initialization order"
echo "   Package: $PKGPATH"

cat > "$TMPDIR/varinit.gno" << EOF
package varinit

import "strconv"

var B = A + 1
var A = 2

func Render(_ string) string {
	return strconv.Itoa(B)
}
EOF

cat > "$TMPDIR/gnomod.toml" << EOF
module = "${PKGPATH}"
gno = "0.9"
EOF

echo -n "   Deploying realm (var B = A+1 declared before var A = 2)... "
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

echo -n "   Querying B (expect 3, A initialized before B)... "
RESULT=$(gnokey query "vm/qeval" \
	-data "${PKGPATH}.Render(\"\")" \
	-remote "$RPC" 2>&1)

if echo "$RESULT" | grep -q '"3"'; then
	echo "✅ PATCHED — B = 3 (A was initialized before B)"
elif echo "$RESULT" | grep -q '"1"'; then
	echo "❌ VULNERABLE — B = 1 (A was 0 when B was initialized)"
	exit 1
else
	echo "⚠️  UNKNOWN OUTPUT"; echo "$RESULT"; exit 1
fi
