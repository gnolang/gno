package main

import (
	"testing"

	signer "github.com/gnolang/gno/tm2/pkg/bft/privval/signer/local"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
