package crypto

import "sync"

var (
	// bech32AddrPrefix defines the Bech32 prefix of an address
	bech32AddrPrefix string

	// bech32PubKeyPrefix defines the Bech32 prefix of a pubkey
	bech32PubKeyPrefix string

	// onceBech32Prefixes ensures the prefixes are set exactly once
	onceBech32Prefixes sync.Once
)

const (
	// Atom in https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	CoinType uint32 = 118

	// BIP44Prefix is the parts of the BIP44 HD path that are fixed by
	// what we used during the fundraiser.
	Bip44DefaultPath = "44'/118'/0'/0/0"
)

func setBech32Defaults() {
	bech32AddrPrefix = "g"
	bech32PubKeyPrefix = "gpub"
}

// Bech32AddrPrefix returns the Bech32 address prefix.
func Bech32AddrPrefix() string {
	onceBech32Prefixes.Do(setBech32Defaults)
	return bech32AddrPrefix
}

// Bech32PubKeyPrefix returns the Bech32 pubkey prefix.
func Bech32PubKeyPrefix() string {
	onceBech32Prefixes.Do(setBech32Defaults)
	return bech32PubKeyPrefix
}

// SetBech32Prefixes sets the Bech32 address and pubkey prefixes.
// This function can only be called once, before any call to the getter functions.
// Subsequent calls panic.
func SetBech32Prefixes(addressPrefix, pubkeyPrefix string) {
	if addressPrefix == "" || pubkeyPrefix == "" {
		panic("bech32 prefixes cannot be empty")
	}

	var executed bool
	onceBech32Prefixes.Do(func() {
		bech32AddrPrefix = addressPrefix
		bech32PubKeyPrefix = pubkeyPrefix
		executed = true
	})
	if !executed {
		panic("bech32 prefixes are already initialized and cannot be changed")
	}
}
