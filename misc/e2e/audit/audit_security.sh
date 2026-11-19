#!/bin/sh
# Targets: fix(gnovm): uint64 overflow at compile time (NEWTENDG-164, 6a6fc4c71)
#          fix(gnovm): iterative stack-overflow recovery (NEWTENDG-182, 3be0408f0)
# Verifies that uint64 constant overflow is caught at compile time and that
# infinite recursion is stopped by the gas limit rather than crashing the node.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=common.sh
. "$SCRIPT_DIR/common.sh"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🛡️  STARTING GNOVM SECURITY AUDIT..."
echo "------------------------------------"

# --- TEST 1: INTEGER OVERFLOW ---
cat > "$TMPDIR/ovf.gno" << 'EOF'
package main
func main() {
    const huge = 18446744073709551615 + 1
    println(huge)
}
EOF

echo -n "🧪 Testing Tier 1 (Integer Overflow, 6a6fc4c71)... "
RESULT_OVF=$(echo "$PASSWORD" | gnokey maketx run \
    -broadcast -remote "$RPC" -chainid "$CHAINID" \
    -gas-fee 1000000ugnot -gas-wanted 2000000 \
    -insecure-password-stdin \
    -home "$GNOKEY_HOME" \
    "$KEY" "$TMPDIR/ovf.gno" 2>&1)

if echo "$RESULT_OVF" | grep -qiE "overflows|cannot use huge"; then
    echo "✅ PATCHED"
else
    echo "❌ VULNERABLE"
    echo "$RESULT_OVF" | grep "Error" | head -n 5
    exit 1
fi

# --- TEST 2: STACK RECURSION ---
cat > "$TMPDIR/kami.gno" << 'EOF'
package main
func main() {
    Recursive()
}
func Recursive() {
    Recursive()
}
EOF

echo -n "🧪 Testing Tier 1 (Stack Recursion, 3be0408f0)... "
RESULT_KAM=$(echo "$PASSWORD" | gnokey maketx run \
    -broadcast -remote "$RPC" -chainid "$CHAINID" \
    -gas-fee 1000000ugnot -gas-wanted 5000000 \
    -insecure-password-stdin \
    -home "$GNOKEY_HOME" \
    "$KEY" "$TMPDIR/kami.gno" 2>&1)

if echo "$RESULT_KAM" | grep -qi "out of gas"; then
    echo "✅ PATCHED (Gas limit hit)"
elif echo "$RESULT_KAM" | grep -qi "stack overflow"; then
    echo "✅ PATCHED (Stack limit hit)"
else
    echo "❌ CRITICAL"
    echo "$RESULT_KAM" | grep "Error" | head -n 5
    exit 1
fi

echo "------------------------------------"
echo "🏁 Audit Complete."
