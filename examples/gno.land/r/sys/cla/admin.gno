package cla

import (
	"gno.land/p/moul/addrset"
	"gno.land/p/nt/ownable"
)

var owner = ownable.NewWithAddress("g1replacemewithmultisig")

// SetRequiredHash updates the required CLA document hash and resets all signatures.
// Only callable by the owner.
// Set to empty string to disable CLA enforcement.
func SetRequiredHash(cur realm, newHash string) {
	owner.AssertOwnedByPrevious()
	requiredHash = newHash
	signatures = addrset.Set{} // reset all signatures
}

// SetCLAURL updates the URL where the CLA document can be found.
// Only callable by the owner.
func SetCLAURL(cur realm, url string) {
	owner.AssertOwnedByPrevious()
	claURL = url
}

// TransferOwnership transfers ownership to a new address.
// Only callable by the current owner.
func TransferOwnership(cur realm, newOwner address) {
	owner.AssertOwnedByPrevious()
	err := owner.TransferOwnership(newOwner)
	if err != nil {
		panic(err)
	}
}
