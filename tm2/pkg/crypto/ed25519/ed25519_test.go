package ed25519_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
)

func TestSignAndValidateEd25519(t *testing.T) {
	t.Parallel()

	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// Mutate the signature, just one bit.
	// TODO: Replace this with a much better fuzzer, tendermint/ed25519/issues/10
	sig[7] ^= byte(0x01)

	assert.False(t, pubKey.VerifyBytes(msg, sig))
}

const (
	privKeySecretGolden = "secret_golden"
	msgGolden           = "msg_golden"
	signedGolden        = "f9d4e6a665dfb6cd7e2fedf0d46a1725472e640a5e93d654ce4caa986e5defd23c8b3af76aa6e39c24c582f0ebee860f66254b29cf6d034ce461ae2773133703"
)

func TestSignAndVerifyGolden(t *testing.T) {
	privKey := ed25519.GenPrivKeyFromSecret([]byte(privKeySecretGolden))
	// pubKey := privKey.PubKey()
	out, err := privKey.Sign([]byte(msgGolden))
	require.NoError(t, err)

	hexOut := hex.EncodeToString(out)
	require.Equal(t, signedGolden, hexOut)
}

func TestPubKeyIsSmallOrder(t *testing.T) {
	t.Parallel()

	// libsodium's ed25519_is_small_order list: order 4, 1, 8, 8, 2, then the
	// non-canonical p (a 0, order 4) and p+1 (a 1, order 1).
	for _, h := range []string{
		"0000000000000000000000000000000000000000000000000000000000000000",
		"0100000000000000000000000000000000000000000000000000000000000000",
		"26e8958fc2b227b045c3f489f2ef98f0d5dfac05d3c63339b13802886d53fc05",
		"c7176a703d4dd84fba3c0b760d10670f2a2053fa2c39ccc64ec7fd7792ac037a",
		"ecffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
		"edffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
		"eeffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff7f",
	} {
		b, err := hex.DecodeString(h)
		require.NoError(t, err)
		for _, sign := range []byte{0, 0x80} {
			var pub ed25519.PubKeyEd25519
			copy(pub[:], b)
			pub[31] |= sign
			assert.True(t, pub.IsSmallOrder(), "%s with sign bit %#x", h, sign)
		}
	}

	for range 100 {
		pub := ed25519.GenPrivKey().PubKey().(ed25519.PubKeyEd25519)
		assert.False(t, pub.IsSmallOrder(), "a key derived from a private key has prime order")
	}
}
