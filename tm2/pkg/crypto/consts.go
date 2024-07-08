package crypto

const (
	// Bech32AddrPrefix defines the Bech32 prefix of an address
	Bech32AddrPrefix = "g"

	// Bech32PubKeyPrefix defines the Bech32 prefix of a pubkey
	Bech32PubKeyPrefix = "gpub"

	// Atom in https://github.com/satoshilabs/slips/blob/master/slip-0044.md
	CoinType uint32 = 118

	// BIP44Prefix is the parts of the BIP44 HD path that are fixed by
	// what we used during the fundraiser.
	Bip44DefaultPath = "44'/118'/0'/0/0"
)
