package crypto

import "sync"

var (
	// bech32AddrPrefix defines the Bech32 prefix of an address
	bech32AddrPrefix = "g"

	// bech32PubKeyPrefix defines the Bech32 prefix of a pubkey
	bech32PubKeyPrefix = "gpub"

	// once guards ensure that setters can only be called once
	onceBech32AddrPrefix   sync.Once
	onceBech32PubKeyPrefix sync.Once
)

const (
	// Atom in https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	CoinType uint32 = 118

	// BIP44Prefix is the parts of the BIP44 HD path that are fixed by
	// what we used during the fundraiser.
	Bip44DefaultPath = "44'/118'/0'/0/0"
)

// GetBech32AddrPrefix returns the Bech32 address prefix.
func GetBech32AddrPrefix() string {
	return bech32AddrPrefix
}

// GetBech32PubKeyPrefix returns the Bech32 pubkey prefix.
func GetBech32PubKeyPrefix() string {
	return bech32PubKeyPrefix
}

// SetBech32AddrPrefix sets the Bech32 address prefix.
// This function can only be called once. Subsequent calls are no-ops.
func SetBech32AddrPrefix(prefix string) {
	onceBech32AddrPrefix.Do(func() {
		bech32AddrPrefix = prefix
	})
}

// SetBech32PubKeyPrefix sets the Bech32 pubkey prefix.
// This function can only be called once. Subsequent calls are no-ops.
func SetBech32PubKeyPrefix(prefix string) {
	onceBech32PubKeyPrefix.Do(func() {
		bech32PubKeyPrefix = prefix
	})
}
