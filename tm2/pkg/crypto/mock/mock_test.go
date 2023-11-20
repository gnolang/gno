package mock_test

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignAndValidateMock(t *testing.T) {
	t.Parallel()

	privKey := mock.PrivKeyMock([]byte{0x01})
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
