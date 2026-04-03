#!/usr/bin/env bash
# Convert a validator bech32 public key (gpub1...) to its gno address (g1...).
#
# Both key types are supported:
#   - secp256k1: regular gnokey secrets-based validator key
#   - ed25519:   gnokms-backed validator key (gnokey validator)
#
# Usage:
#   ./utils/valkey-to-addr.sh <gpub1...>
#
# Example:
#   ./utils/valkey-to-addr.sh gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqdfdtl575xtckdfsjhjwxex2ltwjq7mq36c4y8s4dzcg4gka5pnkq03vsd
#   # => g1u4z9tu4q5838zy07yrd97uu95mkgh4sz5phzsc
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <gpub1...>"
  exit 1
fi

tmpfile=$(mktemp "$REPO_ROOT/valkey-to-addr-XXXXXX.go")
trap 'rm -f "$tmpfile"' EXIT

cat > "$tmpfile" << 'EOF'
package main

import (
	"fmt"
	"os"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	// Register secp256k1 and ed25519 amino types so PubKeyFromBech32 can decode both.
	_ "github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	_ "github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: valkey-to-addr <gpub1...>\n")
		os.Exit(1)
	}

	pubKey, err := crypto.PubKeyFromBech32(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(crypto.AddressToBech32(pubKey.Address()))
}
EOF

go run "$tmpfile" "$1"
