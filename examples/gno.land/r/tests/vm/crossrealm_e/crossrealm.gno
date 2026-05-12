package crossrealm_e

import "chain/runtime"

var (
	balance int64
	owner   address
)

func init() {
	balance = 1000
	SetOwner(address("g1dao_address_here"))
}

// SetOwner is an internal helper that was exported by mistake
// (should be setOwner). Without the pre-mutation readonly check,
// a cross-realm caller could call SetOwner + recover to silently
// hijack ownership in memory, then call TransferToken to steal funds.
func SetOwner(o address) {
	owner = o
}

func GetOwner() address {
	return owner
}

func TransferOwnership(cur realm, o address) {
	if runtime.PreviousRealm().Address() != owner {
		panic("unauthorized")
	}
	owner = o
}

func TransferToken(cur realm) {
	caller := runtime.PreviousRealm().Address()
	if caller != owner {
		panic("unauthorized")
	}
	balance -= 500
	println("===send token to: ", caller)
}
