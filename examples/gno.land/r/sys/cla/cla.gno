package cla

import (
	"chain"
	"chain/runtime"

	"gno.land/p/moul/addrset"
)

const SignedEvent = "CLASigned"

var (
	requiredHash string // SHA256 hash of the CLA document; empty = enforcement disabled
	claURL       string // URL where the CLA document can be found
	signatures   addrset.Set
)

// Sign records a CLA signature for the caller.
// The hash must match the current required hash.
func Sign(cur realm, hash string) {
	if hash != requiredHash {
		panic("hash does not match required CLA hash")
	}

	caller := runtime.PreviousRealm().Address()
	signatures.Add(caller)

	chain.Emit(
		SignedEvent,
		"signer", caller.String(),
		"hash", hash,
	)
}

// HasValidSignature checks if an address has signed the current required CLA.
// Returns true if CLA enforcement is disabled (requiredHash == ""),
// or if the address has signed.
func HasValidSignature(addr address) bool {
	if requiredHash == "" {
		return true
	}
	return signatures.Has(addr)
}
