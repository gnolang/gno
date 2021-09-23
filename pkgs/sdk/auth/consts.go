package auth

import (
	"github.com/gnolang/gno/pkgs/crypto"
)

const (
	// module name
	ModuleName = "auth"

	// StoreKey is string representation of the store key for auth
	StoreKey = "acc"

	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"

	// QuerierRoute is the querier route for acc
	QuerierRoute = StoreKey

	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = "/a/"

	// param key for global account number
	GlobalAccountNumberKey = "globalAccountNumber"
)

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(addr crypto.Address) []byte {
	return append([]byte(AddressStoreKeyPrefix), addr.Bytes()...)
}

// NOTE: do not modify.
// XXX: consider parameterization at the keeper level.
var feeCollector crypto.Address

func FeeCollectorAddress() crypto.Address {
	if feeCollector.IsZero() {
		feeCollector = crypto.AddressFromPreimage([]byte(FeeCollectorName))
	}
	return feeCollector
}
