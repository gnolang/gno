package main

import (
	"path/filepath"
	"testing"

	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
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

		readData, err := readSecretData[types.NodeKey]("")
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

		readData, err := readSecretData[types.NodeKey](path)
		require.Nil(t, readData)

		assert.ErrorContains(t, err, "unable to unmarshal data")
	})

	t.Run("valid data save and read", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		path := filepath.Join(dir, "key.json")
		key := types.GenerateNodeKey()

		require.NoError(t, saveSecretData(key, path))

		readKey, err := readSecretData[types.NodeKey](path)
		require.NoError(t, err)

		assert.Equal(t, key, readKey)
	})
}

func TestCommon_ValidateStateSignature(t *testing.T) {
	t.Parallel()

	t.Run("valid state signature", func(t *testing.T) {
		t.Parallel()

		var (
			key   = signer.GenerateFileKey()
			state = &fstate.FileState{SignBytes: []byte("random data")}
		)

		// Prepare the signature
		signature, err := key.PrivKey.Sign(state.SignBytes)
		require.NoError(t, err)

		state.Signature = signature

		assert.NoError(t, validateValidatorStateSignature(state, key.PubKey))
	})

	t.Run("no state signature", func(t *testing.T) {
		t.Parallel()

		var (
			key   = signer.GenerateFileKey()
			state = &fstate.FileState{}
		)

		assert.NoError(t, validateValidatorStateSignature(state, key.PubKey))
	})

	t.Run("signature mismatch", func(t *testing.T) {
		t.Parallel()

		var (
			key   = signer.GenerateFileKey()
			state = &fstate.FileState{SignBytes: []byte("random data")}
		)

		// Prepare the signature
		signature, err := key.PrivKey.Sign(state.SignBytes)
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

		key := types.GenerateNodeKey()

		assert.NoError(t, validateNodeKey(key))
	})

	t.Run("invalid node key", func(t *testing.T) {
		t.Parallel()

		var key *types.NodeKey = nil

		assert.ErrorIs(t, validateNodeKey(key), errInvalidNodeKey)
	})
}
