#!/bin/sh
# Targets: feat(gnovm)!: move runtime to testing stdlibs (afd7e4808)
# Verifies that importing "runtime" in a production script is rejected.
# The runtime package (GC, MemStats, etc.) has no legitimate on-chain use
# and was removed from production stdlibs. Any deployed realm importing it
# would fail replay after the hardfork.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
. "$SCRIPT_DIR/common.sh"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🧪 afd7e4808 — runtime stdlib removed from production"

cat > "$TMPDIR/runtime.gno" << 'EOF'
package main

import "runtime"

func main() {
	runtime.GC()
}
EOF

echo -n "   Submitting script importing \"runtime\"... "
RESULT=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot \
	-gas-wanted 5000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/runtime.gno" 2>&1)

if echo "$RESULT" | grep -qiE "unknown import|cannot find|not found|unavailable|no package"; then
	echo "✅ PATCHED — runtime import rejected"
elif echo "$RESULT" | grep -q "OK!"; then
	echo "❌ VULNERABLE — runtime.GC() executed in production VM"
	exit 1
else
	echo "⚠️  UNKNOWN OUTPUT"; echo "$RESULT"
	exit 1
fi
