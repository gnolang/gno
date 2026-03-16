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
	// key for gas price
	GasPriceKey = "gasPrice"
	// param key for global account number
	GlobalAccountNumberKey = "globalAccountNumber"

	// SessionStoreKeyInfix is the infix between master address and session address
	SessionStoreKeyInfix = "/s/"
)

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(addr crypto.Address) []byte {
	return append([]byte(AddressStoreKeyPrefix), addr.Bytes()...)
}

// SessionStoreKey returns the store key for a session account: /a/<master>/s/<session>
func SessionStoreKey(master, session crypto.Address) []byte {
	return append(append(AddressStoreKey(master), []byte(SessionStoreKeyInfix)...), session.Bytes()...)
}

// SessionPrefixKey returns the prefix for all sessions of a master: /a/<master>/s/
func SessionPrefixKey(master crypto.Address) []byte {
	return append(AddressStoreKey(master), []byte(SessionStoreKeyInfix)...)
}
