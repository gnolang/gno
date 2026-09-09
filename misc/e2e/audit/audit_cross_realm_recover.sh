#!/bin/sh
# Targets: fix(gnovm): prevent cross-realm state corruption via NameExpr assign+recover (f87249327)
# Verifies that a realm function that writes state then panics causes a full rollback,
# even when the calling script catches the panic with recover().
# Without the fix, the state write was not undone and partial state remained.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
. "$SCRIPT_DIR/common.sh"

SUFFIX=$(date +%s)
PKGPATH="gno.land/r/${KEY_ADDR}/audit/realmrecov${SUFFIX}"
TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🧪 f87249327 — State rollback on panic + recover()"
echo "   Package: $PKGPATH"

cat > "$TMPDIR/realmrecov.gno" << EOF
package realmrecov

import "strconv"

type StateHolder struct {
	value int
}

func (s *StateHolder) set(v int) {
	s.value = v
}

var holder = StateHolder{value: 0}

func SetAndPanic(v int) {
	holder.set(v)
	panic("deliberate panic after state write")
}

func Render(_ string) string {
	return strconv.Itoa(holder.value)
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

cat > "$TMPDIR/recover.gno" << EOF
package main

import r "${PKGPATH}"

func main() {
	defer func() { recover() }()
	r.SetAndPanic(100)
}
EOF

echo -n "   Calling SetAndPanic(100) with recover()... "
CALL=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot -gas-wanted 5000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/recover.gno" 2>&1)
echo "$(echo "$CALL" | grep -oE 'OK!|error' | head -1)"

echo -n "   Querying State (expect 0)... "
RESULT=$(gnokey query "vm/qeval" \
	-data "${PKGPATH}.Render(\"\")" \
	-remote "$RPC" 2>&1)

if echo "$RESULT" | grep -q '"0"'; then
	echo "✅ PATCHED — State rolled back to 0 after panic"
elif echo "$RESULT" | grep -q '"100"'; then
	echo "❌ VULNERABLE — State corrupted to 100 despite panic"
	exit 1
else
	echo "⚠️  UNKNOWN OUTPUT"; echo "$RESULT"; exit 1
fi
