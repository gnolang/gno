package privval

import (
	"os"
	"path"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	fstate "github.com/gnolang/gno/tm2/pkg/bft/privval/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/stretchr/testify/require"
)

const chainID = "chainID"

func TestPubKey(t *testing.T) {
	t.Parallel()

	t.Run("signer succeeded", func(t *testing.T) {
		t.Parallel()

		pv := &PrivValidator{signer: types.NewMockSigner()}
		pk, err := pv.PubKey()
		require.NotNil(t, pk)
		require.NoError(t, err)
	})

	t.Run("signer failed", func(t *testing.T) {
		t.Parallel()

		pv := &PrivValidator{signer: types.NewErroringMockSigner()}
		pk, err := pv.PubKey()
		require.Nil(t, pk)
		require.Error(t, err)
	})
}

func TestSignVote(t *testing.T) {
	t.Parallel()

	t.Run("valid vote signing", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		require.NoError(t, pv.SignVote(chainID, &types.Vote{Type: types.PrecommitType}))
	})

	t.Run("invalid vote type", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		require.Panics(t, func() {
			pv.SignVote(chainID, &types.Vote{Type: types.SignedMsgType(42)})
		})
	})

	t.Run("height, round and step regression", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Vote{
			Height: 4,
			Round:  8,
			Type:   types.PrecommitType,
		}
		require.NoError(t, pv.SignVote(chainID, initialState))

		// Try to sign with an height regression.
		heightRegression := &types.Vote{
			Height: initialState.Height - 1,
			Round:  initialState.Round,
			Type:   initialState.Type,
		}
		require.Error(t, pv.SignVote(chainID, heightRegression))

		// Try to sign with a round regression.
		roundRegression := &types.Vote{
			Height: initialState.Height,
			Round:  initialState.Round - 1,
			Type:   initialState.Type,
		}
		require.Error(t, pv.SignVote(chainID, roundRegression))

		// Try to sign with a step regression.
		stepRegression := &types.Vote{
			Height: initialState.Height,
			Round:  initialState.Round,
			Type:   initialState.Type - 1,
		}
		require.Error(t, pv.SignVote(chainID, stepRegression))
	})

	t.Run("already signed", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Vote{
			Height: 4,
			Round:  8,
			Type:   types.PrecommitType,
		}
		require.NoError(t, pv.SignVote(chainID, initialState))
		require.NotNil(t, initialState.Signature)

		// Try to double sign.
		lastSignature := initialState.Signature
		initialState.Signature = nil
		require.NoError(t, pv.SignVote(chainID, initialState))
		require.Equal(t, lastSignature, initialState.Signature)
	})

	t.Run("already signed with different timestamp", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Vote{
			Height:    4,
			Round:     8,
			Type:      types.PrecommitType,
			Timestamp: tmtime.Now(),
		}
		require.NoError(t, pv.SignVote(chainID, initialState))
		require.NotNil(t, initialState.Signature)

		// Try to double sign.
		lastSignature := initialState.Signature
		lastTimestamp := initialState.Timestamp
		initialState.Signature = nil
		initialState.Timestamp = initialState.Timestamp.Add(42)
		require.NoError(t, pv.SignVote(chainID, initialState))
		require.Equal(t, lastSignature, initialState.Signature)
		require.Equal(t, lastTimestamp, initialState.Timestamp)
	})

	t.Run("same HRS but conflicting sign bytes", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Vote{
			Height: 4,
			Round:  8,
			Type:   types.PrecommitType,
		}
		require.NoError(t, pv.SignVote(chainID, initialState))
		require.NotNil(t, initialState.Signature)

		// Set conflicting sign bytes in state.
		conflictingState := &types.Vote{
			Height: initialState.Height,
			Round:  initialState.Round,
			Type:   types.PrevoteType, // Conflict.
		}
		pv.state.SignBytes = conflictingState.SignBytes(chainID)

		// Try to double sign.
		require.ErrorIs(t, pv.SignVote(chainID, initialState), errSameHRSBadData)
	})

	t.Run("signer Sign error", func(t *testing.T) {
		t.Parallel()

		// Instanciate a new PrivValidator.
		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set erroringMockSigner as PrivValidator signer then try to sign.
		pv.signer = types.NewErroringMockSigner()
		require.ErrorIs(
			t,
			pv.SignVote(chainID, &types.Vote{Type: types.PrecommitType}),
			types.ErrErroringMockSigner,
		)
	})
}

func TestSignProposal(t *testing.T) {
	t.Parallel()

	t.Run("valid proposal signing", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		require.NoError(t, pv.SignProposal(chainID, &types.Proposal{}))
	})

	t.Run("height, round and step regression", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Proposal{
			Height: 4,
			Round:  8,
		}
		require.NoError(t, pv.SignProposal(chainID, initialState))

		// Try to sign with an height regression.
		heightRegression := &types.Proposal{
			Height: initialState.Height - 1,
			Round:  initialState.Round,
		}
		require.Error(t, pv.SignProposal(chainID, heightRegression))

		// Try to sign with a round regression.
		roundRegression := &types.Proposal{
			Height: initialState.Height,
			Round:  initialState.Round - 1,
		}
		require.Error(t, pv.SignProposal(chainID, roundRegression))
	})

	t.Run("already signed", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Proposal{
			Height: 4,
			Round:  8,
		}
		require.NoError(t, pv.SignProposal(chainID, initialState))
		require.NotNil(t, initialState.Signature)

		// Try to double sign.
		lastSignature := initialState.Signature
		initialState.Signature = nil
		require.NoError(t, pv.SignProposal(chainID, initialState))
		require.Equal(t, lastSignature, initialState.Signature)
	})

	t.Run("already signed with different timestamp", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Proposal{
			Height:    4,
			Round:     8,
			Timestamp: tmtime.Now(),
		}
		require.NoError(t, pv.SignProposal(chainID, initialState))
		require.NotNil(t, initialState.Signature)

		// Try to double sign.
		lastSignature := initialState.Signature
		lastTimestamp := initialState.Timestamp
		initialState.Signature = nil
		initialState.Timestamp = initialState.Timestamp.Add(42)
		require.NoError(t, pv.SignProposal(chainID, initialState))
		require.Equal(t, lastSignature, initialState.Signature)
		require.Equal(t, lastTimestamp, initialState.Timestamp)
	})

	t.Run("same HRS but conflicting sign bytes", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set an initial state.
		initialState := &types.Proposal{
			Height: 4,
			Round:  8,
		}
		require.NoError(t, pv.SignProposal(chainID, initialState))
		require.NotNil(t, initialState.Signature)

		// Set conflicting sign bytes in state.
		conflictingState := &types.Proposal{
			Height: initialState.Height + 1, // Conflict.
			Round:  initialState.Round,
		}
		pv.state.SignBytes = conflictingState.SignBytes(chainID)

		// Try to double sign.
		require.ErrorIs(t, pv.SignProposal(chainID, initialState), errSameHRSBadData)
	})

	t.Run("signer Sign error", func(t *testing.T) {
		t.Parallel()

		// Instanciate a new PrivValidator.
		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Set erroringMockSigner as PrivValidator signer then try to sign.
		pv.signer = types.NewErroringMockSigner()
		require.ErrorIs(
			t,
			pv.SignProposal(chainID, &types.Proposal{}),
			types.ErrErroringMockSigner,
		)
	})
}

func TestNewPrivValidator(t *testing.T) {
	t.Parallel()

	t.Run("valid state path", func(t *testing.T) {
		t.Parallel()

		statePath := path.Join(t.TempDir(), "state")
		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)
	})

	t.Run("invalid state path", func(t *testing.T) {
		t.Parallel()

		// Empty state path.
		pv, err := NewPrivValidator(types.NewMockSigner(), "")
		require.Nil(t, pv)
		require.Error(t, err)

		// Create a read-only directory.
		dirPath := path.Join(t.TempDir(), "read-only")
		require.NoError(t, os.Mkdir(dirPath, 0444))

		filePath := path.Join(dirPath, "file")
		pv, err = NewPrivValidator(types.NewMockSigner(), filePath)
		require.Nil(t, pv)
		require.Error(t, err)
	})

	t.Run("signer PubKey getter error", func(t *testing.T) {
		t.Parallel()

		// Create a state on disk.
		statePath := path.Join(t.TempDir(), "state")
		state, err := fstate.NewFileState(statePath)
		require.NotNil(t, state)
		require.NoError(t, err)

		// Update it with sign bytes.
		vote := types.CanonicalVote{Type: types.PrecommitType}
		signBytes, err := amino.MarshalSized(&vote)
		require.NoError(t, err)
		require.NotNil(t, signBytes)
		err = state.Update(0, 0, fstate.StepPrecommit, signBytes, []byte("signature"))
		require.NoError(t, err)
		require.NotNil(t, state.SignBytes)

		pv, err := NewPrivValidator(types.NewErroringMockSigner(), statePath)
		require.Nil(t, pv)
		require.ErrorIs(t, err, types.ErrErroringMockSigner)
	})

	t.Run("invalid state signature", func(t *testing.T) {
		t.Parallel()

		// Create a state on disk.
		statePath := path.Join(t.TempDir(), "state")
		state, err := fstate.NewFileState(statePath)
		require.NotNil(t, state)
		require.NoError(t, err)

		// Update it with invalid sign bytes.
		vote := types.CanonicalVote{Type: types.PrecommitType}
		signBytes, err := amino.MarshalSized(&vote)
		require.NoError(t, err)
		require.NotNil(t, signBytes)
		err = state.Update(0, 0, fstate.StepPrecommit, signBytes, []byte("signature"))
		require.NoError(t, err)
		require.NotNil(t, state.SignBytes)

		pv, err := NewPrivValidator(types.NewMockSigner(), statePath)
		require.Nil(t, pv)
		require.ErrorIs(t, err, errSignatureMismatch)
	})

	t.Run("signer changed", func(t *testing.T) {
		t.Parallel()

		var (
			signer1   = types.NewMockSigner()
			signer2   = types.NewMockSigner()
			statePath = path.Join(t.TempDir(), "state")
		)

		// Instanciate PrivValidator with signer1.
		pv, err := NewPrivValidator(signer1, statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)

		// Sign a vote to update state with signer1 signature.
		vote := &types.Vote{
			Type: types.PrecommitType,
		}
		pv.SignVote(chainID, vote)

		// Instanciate PrivValidator with signer2
		pv, err = NewPrivValidator(signer2, statePath)
		require.Nil(t, pv)
		require.ErrorIs(t, err, errSignatureMismatch)

		// Retry with signer1.
		pv, err = NewPrivValidator(signer1, statePath)
		require.NotNil(t, pv)
		require.NoError(t, err)
	})
}
