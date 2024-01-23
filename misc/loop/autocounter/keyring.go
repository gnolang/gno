package main

import (
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

// MemoryKeyring is an in-memory keyring
type MemoryKeyring struct {
	key crypto.PrivKey
}

func (k MemoryKeyring) GetAddr() crypto.Address {
	return k.key.PubKey().Address()
}

func NewMemoryKeyring(mnemonic string) *MemoryKeyring {
	seed := bip39.NewSeed(mnemonic, "")
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, 0)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)

	//nolint:errcheck // This derivation can never error out, since the path params
	// are always going to be valid
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	key := secp256k1.PrivKeySecp256k1(derivedPriv)

	return &MemoryKeyring{
		key: key,
	}
}
