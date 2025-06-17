package auth

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

const (
	// module name
	ModuleName = "auth"

	// StoreKey is string representation of the store key for auth
	StoreKey = "acc"

	// QuerierRoute is the querier route for acc
	QuerierRoute = StoreKey

	// AddressStoreKeyPrefix prefix for account-by-address store
	AddressStoreKeyPrefix = "/a/"

	// SessionStoreKeyPrefix prefix for account-address-by-pubkey store
	SessionStoreKeyPrefix = "/k/"

	// key for gas price
	GasPriceKey = "gasPrice"

	// param key for global account number
	GlobalAccountNumberKey = "globalAccountNumber"
)

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(addr crypto.Address) []byte {
	return append([]byte(AddressStoreKeyPrefix), addr.Bytes()...)
}

// SessionStoreKey turn a pubkey to key used to get it from the session store
func SessionStoreKey(pubkey crypto.PubKey) []byte {
	return append([]byte(SessionStoreKeyPrefix), pubkey.Bytes()...)
}
