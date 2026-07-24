#!/bin/sh
# Tests replay protection via sequence number enforcement.
# Verifies that rebroadcasting a transaction with an already-consumed sequence
# number is rejected with a sequence mismatch error.
# This is a baseline sanity check that underpins all Tier 1 consensus fixes.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=../audit/common.sh
. "$SCRIPT_DIR/../audit/common.sh"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "🧪 Replay protection — sequence number enforcement"

cat > "$TMPDIR/noop.gno" << 'EOF'
package main

func main() {}
EOF

# Tx 1: normal broadcast, auto-sequence (should succeed)
echo -n "   Tx 1 — normal broadcast... "
TX1=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot -gas-wanted 1000000 \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/noop.gno" 2>&1)

if echo "$TX1" | grep -q "OK!"; then
	echo "OK"
else
	echo "FAILED (unexpected)"; echo "$TX1"; exit 1
fi

# Derive consumed sequence from the account state
SEQ_INFO=$(gnokey query "auth/accounts/${KEY_ADDR}" -remote "$RPC" 2>&1)
CURRENT_SEQ=$(echo "$SEQ_INFO" | grep -oE '"sequence":"[0-9]+"' | grep -oE '[0-9]+$')
if [ -z "$CURRENT_SEQ" ] || [ "$CURRENT_SEQ" -eq 0 ]; then
	REPLAY_SEQ=0
else
	REPLAY_SEQ=$((CURRENT_SEQ - 1))
fi
echo "   Current sequence: $CURRENT_SEQ — replaying with sequence: $REPLAY_SEQ"

# Tx 2: replay with the already-used sequence number (must be rejected)
echo -n "   Tx 2 — replay at sequence $REPLAY_SEQ (expect rejection)... "
TX2=$(echo "$PASSWORD" | gnokey maketx run \
	-gas-fee 1000000ugnot -gas-wanted 1000000 \
	-sequence "$REPLAY_SEQ" \
	-broadcast -chainid "$CHAINID" -remote "$RPC" \
	-insecure-password-stdin \
	-home "$GNOKEY_HOME" \
	"$KEY" "$TMPDIR/noop.gno" 2>&1)

if echo "$TX2" | grep -qiE "sequence|wrong nonce|invalid sequence|account sequence|mempool"; then
	echo "✅ PROTECTED — replay rejected"
elif echo "$TX2" | grep -q "OK!"; then
	echo "❌ VULNERABLE — replay accepted"
	exit 1
else
	echo "⚠️  UNKNOWN OUTPUT"; echo "$TX2"; exit 1
fi
