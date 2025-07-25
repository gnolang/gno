package state

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStepToVote(t *testing.T) {
	t.Parallel()

	t.Run("valid step conversion", func(t *testing.T) {
		t.Parallel()

		require.NotPanics(t, func() { VoteTypeToStep(types.PrevoteType) })
		assert.NotPanics(t, func() { VoteTypeToStep(types.PrecommitType) })
	})

	t.Run("invalid step conversion", func(t *testing.T) {
		t.Parallel()

		require.Panics(t, func() { VoteTypeToStep(types.ProposalType) })
		assert.Panics(t, func() { VoteTypeToStep(4) })
	})
}

func TestCheckHRS(t *testing.T) {
	t.Parallel()

	fs := &FileState{
		Height:    1,
		Round:     2,
		Step:      3,
		SignBytes: []byte("Not nil"),
		Signature: []byte("Not nil"),
	}

	t.Run("reusable HRS", func(t *testing.T) {
		t.Parallel()

		reusable, err := fs.CheckHRS(fs.Height, fs.Round, fs.Step)
		require.True(t, reusable)
		assert.NoError(t, err)
	})

	t.Run("height regression", func(t *testing.T) {
		t.Parallel()

		reusable, err := fs.CheckHRS(fs.Height-1, fs.Round, fs.Step)
		require.False(t, reusable)
		assert.ErrorIs(t, err, errHeightRegression)
	})

	t.Run("height progression", func(t *testing.T) {
		t.Parallel()

		reusable, err := fs.CheckHRS(fs.Height+1, fs.Round, fs.Step)
		require.False(t, reusable)
		assert.NoError(t, err)
	})

	t.Run("round regression", func(t *testing.T) {
		t.Parallel()

		reusable, err := fs.CheckHRS(fs.Height, fs.Round-1, fs.Step)
		require.False(t, reusable)
		assert.ErrorIs(t, err, errRoundRegression)
	})

	t.Run("round progression", func(t *testing.T) {
		t.Parallel()

		reusable, err := fs.CheckHRS(fs.Height, fs.Round+1, fs.Step)
		require.False(t, reusable)
		assert.NoError(t, err)
	})

	t.Run("step regression", func(t *testing.T) {
		t.Parallel()

		reusable, err := fs.CheckHRS(fs.Height, fs.Round, fs.Step-1)
		require.False(t, reusable)
		assert.ErrorIs(t, err, errStepRegression)
	})

	t.Run("step progression", func(t *testing.T) {
		t.Parallel()

		reusable, err := fs.CheckHRS(fs.Height, fs.Round, fs.Step+1)
		require.False(t, reusable)
		assert.NoError(t, err)
	})

	t.Run("sign bytes not set", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{SignBytes: nil}
		reusable, err := fs.CheckHRS(fs.Height, fs.Round, fs.Step)
		require.False(t, reusable)
		assert.ErrorIs(t, err, errNoSignBytes)
	})

	t.Run("signature not set", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{SignBytes: []byte("Not nil"), Signature: nil}
		assert.Panics(t, func() { fs.CheckHRS(fs.Height, fs.Round, fs.Step) })
	})
}

func TestCheckVotesOnlyDifferByTimestamp(t *testing.T) {
	t.Parallel()

	generateVoteSignBytes := func(timestamp time.Time, height int64) []byte {
		t.Helper()

		signBytes, _ := amino.MarshalSized(types.CanonicalVote{
			Height:    height,
			Timestamp: timestamp,
		})

		return signBytes
	}

	t.Run("invalid state sign bytes", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{SignBytes: []byte("invalid sign bytes")}
		assert.Panics(t, func() { fs.CheckVotesOnlyDifferByTimestamp(nil) })
	})

	t.Run("invalid parameter sign bytes", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{SignBytes: generateVoteSignBytes(tmtime.Now(), 1)}
		assert.Panics(t, func() { fs.CheckVotesOnlyDifferByTimestamp([]byte("invalid sign bytes")) })
	})

	t.Run("differ by timestamp", func(t *testing.T) {
		t.Parallel()

		lastTime := tmtime.Now()
		fs := &FileState{SignBytes: generateVoteSignBytes(lastTime, 1)}
		retTime, timeDiffOnly := fs.CheckVotesOnlyDifferByTimestamp(generateVoteSignBytes(lastTime.Add(1), 1))
		require.Equal(t, retTime, lastTime)
		assert.True(t, timeDiffOnly)
	})

	t.Run("differ by height", func(t *testing.T) {
		t.Parallel()

		lastTime := tmtime.Now()
		fs := &FileState{SignBytes: generateVoteSignBytes(lastTime, 1)}
		retTime, timeDiffOnly := fs.CheckVotesOnlyDifferByTimestamp(generateVoteSignBytes(lastTime, 2))
		require.Equal(t, retTime, lastTime)
		assert.False(t, timeDiffOnly)
	})
}

func TestCheckProposalsOnlyDifferByTimestamp(t *testing.T) {
	t.Parallel()

	generateProposalSignBytes := func(timestamp time.Time, height int64) []byte {
		t.Helper()

		signBytes, _ := amino.MarshalSized(types.CanonicalProposal{
			Height:    height,
			Timestamp: timestamp,
		})

		return signBytes
	}

	t.Run("invalid state sign bytes", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{SignBytes: []byte("invalid sign bytes")}
		assert.Panics(t, func() { fs.CheckProposalsOnlyDifferByTimestamp(nil) })
	})

	t.Run("invalid parameter sign bytes", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{SignBytes: generateProposalSignBytes(tmtime.Now(), 1)}
		assert.Panics(t, func() { fs.CheckProposalsOnlyDifferByTimestamp([]byte("invalid sign bytes")) })
	})

	t.Run("differ by timestamp", func(t *testing.T) {
		t.Parallel()

		lastTime := tmtime.Now()
		fs := &FileState{SignBytes: generateProposalSignBytes(lastTime, 1)}
		retTime, timeDiffOnly := fs.CheckProposalsOnlyDifferByTimestamp(generateProposalSignBytes(lastTime.Add(1), 1))
		require.Equal(t, retTime, lastTime)
		assert.True(t, timeDiffOnly)
	})

	t.Run("differ by height", func(t *testing.T) {
		t.Parallel()

		lastTime := tmtime.Now()
		fs := &FileState{SignBytes: generateProposalSignBytes(lastTime, 1)}
		retTime, timeDiffOnly := fs.CheckProposalsOnlyDifferByTimestamp(generateProposalSignBytes(lastTime, 2))
		require.Equal(t, retTime, lastTime)
		assert.False(t, timeDiffOnly)
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("update without change", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "new")

		// Generate a valid FileState first.
		fs, err := GeneratePersistedFileState(filePath)
		require.NotNil(t, fs)
		require.NoError(t, err)

		// Update without change.
		err = fs.Update(fs.Height, fs.Round, fs.Step, fs.SignBytes, fs.Signature)
		require.NoError(t, err)

		// Load from disk and compare.
		loaded, err := LoadFileState(filePath)
		require.NotNil(t, loaded)
		require.NoError(t, err)
		assert.Equal(t, fs, loaded)
	})

	t.Run("update with change", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "new")

		// Generate a valid FileState first.
		fs, err := GeneratePersistedFileState(filePath)
		require.NotNil(t, fs)
		require.NoError(t, err)

		// Update with change.
		err = fs.Update(fs.Height+1, fs.Round, fs.Step, fs.SignBytes, fs.Signature)
		require.NoError(t, err)

		// Load from disk and compare.
		loaded, err := LoadFileState(filePath)
		require.NotNil(t, loaded)
		require.NoError(t, err)
		assert.Equal(t, fs, loaded)
	})

	t.Run("update with invalid change", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "new")

		// Generate a valid FileState first.
		fs, err := GeneratePersistedFileState(filePath)
		require.NotNil(t, fs)
		require.NoError(t, err)

		// Update with invalid step change.
		err = fs.Update(fs.Height, fs.Round, StepPrecommit+1, fs.SignBytes, fs.Signature)
		assert.ErrorIs(t, err, errInvalidSignStateStep)
	})
}

func TestStringer(t *testing.T) {
	t.Parallel()

	state := &FileState{
		Height: 42,
		Round:  24,
		Step:   StepPrecommit,
	}

	assert.Equal(t, state.String(), "{H: 42, R: 24, S: 3}")
}

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("invalid sign state height", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{Height: -1}
		assert.ErrorIs(t, fs.validate(), errInvalidSignStateHeight)
	})

	t.Run("invalid sign state round", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{Round: -1}
		assert.ErrorIs(t, fs.validate(), errInvalidSignStateRound)
	})

	t.Run("invalid sign state step", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{Step: StepPrecommit + 1}
		assert.ErrorIs(t, fs.validate(), errInvalidSignStateStep)
	})

	t.Run("invalid sign state sign bytes", func(t *testing.T) {
		t.Parallel()

		// Test totally invalid sign bytes.
		fs := &FileState{
			Height:    1,
			Round:     2,
			SignBytes: []byte("invalid sign bytes"),
		}
		require.ErrorIs(t, fs.validate(), errInvalidSignStateSignBytes)
		fs.Step = StepPrevote
		require.ErrorIs(t, fs.validate(), errInvalidSignStateSignBytes)
		fs.Step = StepPropose
		require.ErrorIs(t, fs.validate(), errInvalidSignStateSignBytes)

		// Test valid vote sign bytes with height mismatch.
		var err error
		fs.Step = StepPrecommit
		vote := types.CanonicalVote{
			Type:   types.PrecommitType,
			Height: fs.Height + 1,
			Round:  int64(fs.Round),
		}
		fs.SignBytes, err = amino.MarshalSized(&vote)
		require.NoError(t, err)
		require.ErrorIs(t, fs.validate(), errInvalidSignStateSignBytes)

		// Test valid vote sign bytes with round mismatch.
		vote.Height = fs.Height
		vote.Round = int64(fs.Round) + 1
		fs.SignBytes, err = amino.MarshalSized(&vote)
		require.NoError(t, err)
		require.ErrorIs(t, fs.validate(), errInvalidSignStateSignBytes)

		// Test valid vote sign bytes with step mismatch.
		vote.Round = int64(fs.Round)
		vote.Type = types.PrevoteType
		fs.SignBytes, err = amino.MarshalSized(&vote)
		require.NoError(t, err)
		require.ErrorIs(t, fs.validate(), errInvalidSignStateSignBytes)

		// Test valid proposal sign bytes with height mismatch.
		proposal := types.CanonicalProposal{
			Type:   types.ProposalType,
			Height: fs.Height + 1,
			Round:  int64(fs.Round),
		}
		fs.Step = StepPropose
		fs.SignBytes, err = amino.MarshalSized(&proposal)
		require.NoError(t, err)
		assert.ErrorIs(t, fs.validate(), errInvalidSignStateSignBytes)
	})

	t.Run("valid sign state sign bytes", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{
			Height:   1,
			Round:    2,
			filePath: "valid path",
		}

		// Test valid vote sign bytes.
		var err error
		vote := types.CanonicalVote{
			Type:   types.PrecommitType,
			Height: fs.Height,
			Round:  int64(fs.Round),
		}
		fs.Step = StepPrecommit
		fs.SignBytes, err = amino.MarshalSized(&vote)
		require.NoError(t, err)
		require.NoError(t, fs.validate())

		// Test valid proposal sign bytes.
		proposal := types.CanonicalProposal{
			Type:   types.ProposalType,
			Height: fs.Height,
			Round:  int64(fs.Round),
		}
		fs.Step = StepPropose
		fs.SignBytes, err = amino.MarshalSized(&proposal)
		require.NoError(t, err)
		assert.NoError(t, fs.validate())
	})

	t.Run("signature set but no sign bytes", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{
			SignBytes: nil,
			Signature: []byte("signature that should be nil"),
		}
		assert.ErrorIs(t, fs.validate(), errSignatureShouldNotBeSet)
	})

	t.Run("filepath not set", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{}
		assert.ErrorIs(t, fs.validate(), errFilePathNotSet)
	})

	t.Run("valid file state", func(t *testing.T) {
		t.Parallel()

		fs := &FileState{filePath: "valid path"}
		assert.NoError(t, fs.validate())
	})
}

func TestSave(t *testing.T) {
	t.Parallel()

	t.Run("empty file path", func(t *testing.T) {
		t.Parallel()

		fs, err := GeneratePersistedFileState("")
		require.Nil(t, fs)
		assert.Error(t, err)
	})

	t.Run("read-only file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only directory.
		dirPath := path.Join(t.TempDir(), "read-only")
		err := os.Mkdir(dirPath, 0o444)
		require.NoError(t, err)

		filePath := path.Join(dirPath, "file")
		fs, err := GeneratePersistedFileState(filePath)
		require.Nil(t, fs)
		assert.Error(t, err)
	})

	t.Run("read-write file path", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "writable")
		fs, err := GeneratePersistedFileState(filePath)
		require.NotNil(t, fs)
		assert.NoError(t, err)
	})
}

func TestLoadFileState(t *testing.T) {
	t.Parallel()

	t.Run("valid file state", func(t *testing.T) {
		t.Parallel()

		// Generate a valid random file state on disk.
		filePath := path.Join(t.TempDir(), "valid")
		fs, err := GeneratePersistedFileState(filePath)
		require.NoError(t, err)

		// Load the file state from disk.
		loaded, err := LoadFileState(filePath)
		require.NoError(t, err)

		// Compare the loaded file state with the original.
		assert.Equal(t, fs, loaded)
	})

	t.Run("non-existent file path", func(t *testing.T) {
		t.Parallel()

		fs, err := LoadFileState("non-existent")
		require.Nil(t, fs)
		assert.Error(t, err)
	})

	t.Run("invalid file state", func(t *testing.T) {
		t.Parallel()

		// Create a file with invalid FileState JSON.
		filePath := path.Join(t.TempDir(), "invalid")
		os.WriteFile(filePath, []byte(`{height:"invalid"}`), 0o644)

		fs, err := LoadFileState(filePath)
		require.Nil(t, fs)
		require.Error(t, err)

		// Generate a valid FileState first.
		fs, err = GeneratePersistedFileState(filePath)
		require.NotNil(t, fs)
		require.NoError(t, err)

		// Make its address invalid then persist it to disk.
		fs.Step = StepPrecommit + 1
		jsonBytes, err := amino.MarshalJSONIndent(fs, "", "  ")
		require.NoError(t, err)
		require.NoError(t, osm.WriteFileAtomic(fs.filePath, jsonBytes, 0o600))

		fs, err = LoadFileState(filePath)
		require.Nil(t, fs)
		assert.ErrorIs(t, err, errInvalidSignStateStep)
	})
}

func TestNewFileState(t *testing.T) {
	t.Parallel()

	t.Run("genetate new state", func(t *testing.T) {
		t.Parallel()

		filePath := path.Join(t.TempDir(), "new")
		fs, err := LoadOrMakeFileState(filePath)
		require.NotNil(t, fs)
		assert.NoError(t, err)
	})

	t.Run("load existing state", func(t *testing.T) {
		t.Parallel()

		// Generate a valid random file state on disk.
		filePath := path.Join(t.TempDir(), "existing")
		fs, err := GeneratePersistedFileState(filePath)
		require.NoError(t, err)

		// Load it using NewFileState.
		loaded, err := LoadOrMakeFileState(filePath)
		require.NotNil(t, loaded)
		require.NoError(t, err)

		// Compare the loaded file state with the original.
		assert.Equal(t, fs, loaded)
	})

	t.Run("read-only file path", func(t *testing.T) {
		t.Parallel()

		// Create a read-only directory.
		dirPath := path.Join(t.TempDir(), "read-only")
		err := os.Mkdir(dirPath, 0o444)
		require.NoError(t, err)

		filePath := path.Join(dirPath, "file")
		fs, err := LoadOrMakeFileState(filePath)
		require.Nil(t, fs)
		assert.Error(t, err)
	})
}
