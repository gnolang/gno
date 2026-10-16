#!/bin/sh
# Targets: fix(gnovm): reject chan type at preprocess/runtime (4bcd9828e)
# Verifies that chan types are rejected at preprocess time (before execution).
# Without the fix, deployment succeeded and the node panicked only at runtime
# when the channel was actually used.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
. "$SCRIPT_DIR/common.sh"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🧪 4bcd9828e — chan type rejection at preprocess"

cat > "$TMPDIR/chan.gno" << 'EOF'
package main

func main() {
	ch := make(chan int, 1)
	ch <- 42
}
EOF

echo -n "   Submitting script with chan type... "
RESULT=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot \
	-gas-wanted 5000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/chan.gno" 2>&1)

if echo "$RESULT" | grep -q "OK!"; then
	echo "FAIL: chan type accepted by the VM (VULNERABLE)"
	exit 1
else
	echo "PASS: chan type rejected (PATCHED)"
fi
