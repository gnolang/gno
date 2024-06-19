package main

import (
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommon_SaveReadData(t *testing.T) {
	t.Parallel()

	t.Run("invalid data save path", func(t *testing.T) {
		t.Parallel()

		assert.ErrorContains(
			t,
			saveSecretData(nil, ""),
			"unable to save data to path",
		)
	})

	t.Run("invalid data read path", func(t *testing.T) {
		t.Parallel()

		readData, err := readSecretData[p2p.NodeKey]("")
		assert.Nil(t, readData)

		assert.ErrorContains(
			t,
			err,
			"unable to read data",
		)
	})

	t.Run("invalid data read", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "key.json")

		require.NoError(t, saveSecretData("totally valid key", path))

		readData, err := readSecretData[p2p.NodeKey](path)
		require.Nil(t, readData)

		assert.ErrorContains(t, err, "unable to unmarshal data")
	})

	t.Run("valid data save and read", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "key.json")
		key := generateNodeKey()

		require.NoError(t, saveSecretData(key, path))

		readKey, err := readSecretData[p2p.NodeKey](path)
		require.NoError(t, err)

		assert.Equal(t, key, readKey)
	})
}

func TestCommon_ValidateValidatorKey(t *testing.T) {
	t.Parallel()

	t.Run("valid validator key", func(t *testing.T) {
		t.Parallel()

		key := generateValidatorPrivateKey()

		assert.NoError(t, validateValidatorKey(key))
	})

	t.Run("invalid private key", func(t *testing.T) {
		t.Parallel()

		key := generateValidatorPrivateKey()
		key.PrivKey = nil

		assert.ErrorIs(t, validateValidatorKey(key), errInvalidPrivateKey)
	})

	t.Run("public key mismatch", func(t *testing.T) {
		t.Parallel()

		key := generateValidatorPrivateKey()
		key.PubKey = nil

		assert.ErrorIs(t, validateValidatorKey(key), errPublicKeyMismatch)
	})

	t.Run("address mismatch", func(t *testing.T) {
		t.Parallel()

		key := generateValidatorPrivateKey()
		key.Address = crypto.Address{} // zero address

		assert.ErrorIs(t, validateValidatorKey(key), errAddressMismatch)
	})
}

func TestCommon_ValidateValidatorState(t *testing.T) {
	t.Parallel()

	t.Run("valid validator state", func(t *testing.T) {
		t.Parallel()

		state := generateLastSignValidatorState()

		assert.NoError(t, validateValidatorState(state))
	})

	t.Run("invalid step", func(t *testing.T) {
		t.Parallel()

		state := generateLastSignValidatorState()
		state.Step = -1

		assert.ErrorIs(t, validateValidatorState(state), errInvalidSignStateStep)
	})

	t.Run("invalid height", func(t *testing.T) {
		t.Parallel()

		state := generateLastSignValidatorState()
		state.Height = -1

		assert.ErrorIs(t, validateValidatorState(state), errInvalidSignStateHeight)
	})

	t.Run("invalid round", func(t *testing.T) {
		t.Parallel()

		state := generateLastSignValidatorState()
		state.Round = -1

		assert.ErrorIs(t, validateValidatorState(state), errInvalidSignStateRound)
	})
}

func TestCommon_ValidateStateSignature(t *testing.T) {
	t.Parallel()

	t.Run("valid state signature", func(t *testing.T) {
		t.Parallel()

		var (
			key   = generateValidatorPrivateKey()
			state = generateLastSignValidatorState()

			signData = []byte("random data")
		)

		// Prepare the signature
		signature, err := key.PrivKey.Sign(signData)
		require.NoError(t, err)

		state.Signature = signature
		state.SignBytes = signData

		assert.NoError(t, validateValidatorStateSignature(state, key.PubKey))
	})

	t.Run("no state signature", func(t *testing.T) {
		t.Parallel()

		var (
			key   = generateValidatorPrivateKey()
			state = generateLastSignValidatorState()
		)

		assert.NoError(t, validateValidatorStateSignature(state, key.PubKey))
	})

	t.Run("signature values missing, sign bytes", func(t *testing.T) {
		t.Parallel()

		var (
			key   = generateValidatorPrivateKey()
			state = generateLastSignValidatorState()
		)

		state.Signature = []byte("signature")

		assert.ErrorIs(
			t,
			validateValidatorStateSignature(state, key.PubKey),
			errSignatureValuesMissing,
		)
	})

	t.Run("signature values missing, signature", func(t *testing.T) {
		t.Parallel()

		var (
			key   = generateValidatorPrivateKey()
			state = generateLastSignValidatorState()
		)

		state.SignBytes = []byte("signature")

		assert.ErrorIs(
			t,
			validateValidatorStateSignature(state, key.PubKey),
			errSignatureValuesMissing,
		)
	})

	t.Run("signature mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			key   = generateValidatorPrivateKey()
			state = generateLastSignValidatorState()

			signData = []byte("random data")
		)

		// Prepare the signature
		signature, err := key.PrivKey.Sign(signData)
		require.NoError(t, err)

		state.Signature = signature
		state.SignBytes = []byte("something different")

		assert.ErrorIs(
			t,
			validateValidatorStateSignature(state, key.PubKey),
			errSignatureMismatch,
		)
	})
}

func TestCommon_ValidateNodeKey(t *testing.T) {
	t.Parallel()

	t.Run("valid node key", func(t *testing.T) {
		t.Parallel()

		key := generateNodeKey()

		assert.NoError(t, validateNodeKey(key))
	})

	t.Run("invalid node key", func(t *testing.T) {
		t.Parallel()

		key := generateNodeKey()
		key.PrivKey = nil

		assert.ErrorIs(t, validateNodeKey(key), errInvalidNodeKey)
	})
}
