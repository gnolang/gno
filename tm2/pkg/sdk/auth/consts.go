package auth

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

const (
	// module name
	ModuleName = "auth"

	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"

	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = "/a/"
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
