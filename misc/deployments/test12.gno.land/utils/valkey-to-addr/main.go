// valkey-to-addr converts a validator bech32 public key (gpub1...) to its
// corresponding gno address (g1...).
//
// Both secp256k1 (regular gnokey secrets) and ed25519 (gnokms validator) key
// types are supported — amino decoding identifies the type automatically.
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
