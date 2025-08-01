package common

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/bip39"
	"github.com/gnolang/gno/tm2/pkg/crypto/hd"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys/client"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

// DummyKey generates a random public key,
// and returns the key info
func DummyKey(t *testing.T) crypto.PubKey {
	t.Helper()

	mnemonic, err := client.GenerateMnemonic(256)
	require.NoError(t, err)

	seed := bip39.NewSeed(mnemonic, "")

	return generateKeyFromSeed(seed, 0).PubKey()
}

// generateKeyFromSeed generates a private key from
// the provided seed and index
func generateKeyFromSeed(seed []byte, index uint32) crypto.PrivKey {
	pathParams := hd.NewFundraiserParams(0, crypto.CoinType, index)

	masterPriv, ch := hd.ComputeMastersFromSeed(seed)

	//nolint:errcheck // This derivation can never error out, since the path params
	// are always going to be valid
	derivedPriv, _ := hd.DerivePrivateKeyForPath(masterPriv, ch, pathParams.String())

	return secp256k1.PrivKeySecp256k1(derivedPriv)
}

// DummyKeys generates random keys for testing
func DummyKeys(t *testing.T, count int) []crypto.PubKey {
	t.Helper()

	dummyKeys := make([]crypto.PubKey, count)

	for i := 0; i < count; i++ {
		dummyKeys[i] = DummyKey(t)
	}

	return dummyKeys
}
