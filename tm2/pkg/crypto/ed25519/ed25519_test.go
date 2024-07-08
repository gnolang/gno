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
