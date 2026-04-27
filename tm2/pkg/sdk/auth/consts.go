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

	// AddressStoreKeyPrefix prefix for account-by-address store.
	//
	// IMPORTANT: Session accounts are also stored under this prefix,
	// at /a/<master>/s/<session>. This means a PrefixIterator on "/a/"
	// returns BOTH regular accounts and session accounts. Use
	// AccountStoreKeyLen to filter: regular account keys are exactly
	// len("/a/") + crypto.AddressSize bytes. Anything longer is a
	// session sub-key. See IterateAccounts for the canonical filter.
	AddressStoreKeyPrefix = "/a/"

	// AccountStoreKeyLen is the exact byte length of a regular account
	// store key: len("/a/") + crypto.AddressSize. Keys under "/a/" that
	// are longer than this are session sub-keys, not regular accounts.
	// Used by IterateAccounts to skip session accounts during iteration.
	AccountStoreKeyLen = len(AddressStoreKeyPrefix) + crypto.AddressSize

	// key for gas price
	GasPriceKey = "gasPrice"
	// param key for global account number
	GlobalAccountNumberKey = "globalAccountNumber"

	// SessionStoreKeyInfix separates master and session addresses in
	// session account keys. The full key format is:
	//
	//   /a/<master 20 bytes>/s/<session 20 bytes>
	//
	// This nests sessions under the "/a/" prefix so they share IAVL
	// tree nodes with the master account (cheap second read). The "/s/"
	// infix acts as a visual delimiter for debugging raw store dumps.
	//
	// IMPORTANT: Because sessions share the "/a/" prefix, any code
	// that iterates "/a/" MUST filter by key length to exclude session
	// keys. See AccountStoreKeyLen and IterateAccounts.
	SessionStoreKeyInfix = "/s/"
)

// AddressStoreKey returns the store key for a regular account: /a/<addr>.
// The resulting key is exactly AccountStoreKeyLen bytes.
func AddressStoreKey(addr crypto.Address) []byte {
	return append([]byte(AddressStoreKeyPrefix), addr.Bytes()...)
}

// SessionStoreKey returns the store key for a session account:
// /a/<master>/s/<session>. This key is longer than AccountStoreKeyLen,
// which is how IterateAccounts distinguishes sessions from regular accounts.
func SessionStoreKey(master, session crypto.Address) []byte {
	return append(append(AddressStoreKey(master), []byte(SessionStoreKeyInfix)...), session.Bytes()...)
}

// SessionPrefixKey returns the prefix for all sessions of a master:
// /a/<master>/s/. Used for prefix iteration and RevokeAll.
func SessionPrefixKey(master crypto.Address) []byte {
	return append(AddressStoreKey(master), []byte(SessionStoreKeyInfix)...)
}
