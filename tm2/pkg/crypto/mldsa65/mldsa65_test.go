package mldsa65_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/mldsa65"
)

func TestSignAndVerify(t *testing.T) {
	t.Parallel()

	privKey := mldsa65.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(128)
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)
	require.Len(t, sig, mldsa65.SignatureSize)
	require.Len(t, pubKey.(mldsa65.PubKeyMLDSA65), mldsa65.PubKeySize)

	assert.True(t, pubKey.VerifyBytes(msg, sig))

	// A flipped bit in the signature fails.
	badSig := bytes.Clone(sig)
	badSig[7] ^= 0x01
	assert.False(t, pubKey.VerifyBytes(msg, badSig))

	// A wrong length fails without reaching the verifier.
	assert.False(t, pubKey.VerifyBytes(msg, sig[:len(sig)-1]))
	assert.False(t, pubKey.VerifyBytes(msg, nil))

	// Another key does not verify it.
	assert.False(t, mldsa65.GenPrivKey().PubKey().VerifyBytes(msg, sig))
}

func TestWrongMessageFails(t *testing.T) {
	t.Parallel()

	privKey := mldsa65.GenPrivKey()
	sig, err := privKey.Sign([]byte("pay alice 10"))
	require.NoError(t, err)

	pubKey := privKey.PubKey()
	assert.True(t, pubKey.VerifyBytes([]byte("pay alice 10"), sig))
	assert.False(t, pubKey.VerifyBytes([]byte("pay alice 11"), sig))
	assert.False(t, pubKey.VerifyBytes(nil, sig))
}

func TestGenPrivKeyFromSecretIsDeterministic(t *testing.T) {
	t.Parallel()

	a := mldsa65.GenPrivKeyFromSecret([]byte("secret_golden"))
	b := mldsa65.GenPrivKeyFromSecret([]byte("secret_golden"))
	c := mldsa65.GenPrivKeyFromSecret([]byte("secret_other"))

	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
	assert.True(t, a.PubKey().Equals(b.PubKey()))
	assert.False(t, a.PubKey().Equals(c.PubKey()))
	assert.Equal(t, a.PubKey().Address(), b.PubKey().Address())

	// Signing is deterministic too.
	msg := []byte("msg_golden")
	sigA, err := a.Sign(msg)
	require.NoError(t, err)
	sigB, err := b.Sign(msg)
	require.NoError(t, err)
	assert.Equal(t, sigA, sigB)
	assert.True(t, b.PubKey().VerifyBytes(msg, sigA))
}

func TestAddressIsTruncatedSHA256(t *testing.T) {
	t.Parallel()

	pub := mldsa65.GenPrivKey().PubKey().(mldsa65.PubKeyMLDSA65)
	want := crypto.Sha256(pub[:])[:crypto.AddressSize]
	assert.Equal(t, want, pub.Address().Bytes())
}

func TestAminoRoundTripThroughInterface(t *testing.T) {
	t.Parallel()

	privKey := mldsa65.GenPrivKey()
	pubKey := privKey.PubKey()

	// Binary, Any-wrapped, through crypto.PubKey.
	var decodedPub crypto.PubKey
	require.NoError(t, amino.UnmarshalAny(pubKey.Bytes(), &decodedPub))
	require.IsType(t, mldsa65.PubKeyMLDSA65{}, decodedPub)
	assert.True(t, pubKey.Equals(decodedPub))

	// Binary, Any-wrapped, through crypto.PrivKey.
	var decodedPriv crypto.PrivKey
	require.NoError(t, amino.UnmarshalAny(privKey.Bytes(), &decodedPriv))
	assert.True(t, privKey.Equals(decodedPriv))

	// JSON through crypto.PubKey.
	bz, err := amino.MarshalJSONAny(pubKey)
	require.NoError(t, err)
	var jsonPub crypto.PubKey
	require.NoError(t, amino.UnmarshalJSON(bz, &jsonPub))
	assert.True(t, pubKey.Equals(jsonPub))

	// Bech32 string round trip.
	parsed, err := crypto.PubKeyFromBech32(pubKey.String())
	require.NoError(t, err)
	assert.True(t, pubKey.Equals(parsed))

	// The decoded key still verifies.
	msg := []byte("round trip")
	sig, err := privKey.Sign(msg)
	require.NoError(t, err)
	assert.True(t, decodedPub.VerifyBytes(msg, sig))
}
