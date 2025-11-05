package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/bft/mempool/mock"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
	"github.com/gnolang/gno/tm2/pkg/log"
)

const validationTestsStopHeight int64 = 10

func TestValidateBlockHeader(t *testing.T) {
	t.Parallel()

	proxyApp := newTestApp()
	require.NoError(t, proxyApp.Start())
	defer proxyApp.Stop()

	state, stateDB, privVals := makeState(3, 1)
	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	lastCommit := types.NewCommit(types.BlockID{}, nil)

	validHash := tmhash.Sum([]byte("this hash is valid"))
	wrongHash := tmhash.Sum([]byte("this hash is wrong"))

	wrongAddress := ed25519.GenPrivKey().PubKey().Address()
	invalidAddress := crypto.Address{}

	// Manipulation of any header field causes failure.
	testCases := []struct {
		name          string
		malleateBlock func(block *types.Block)
		expectedError string
	}{
		{
			"BlockVersion wrong",
			func(block *types.Block) { block.Version += "-wrong" },
			"wrong Block.Header.Version",
		},
		{
			"AppVersion wrong",
			func(block *types.Block) { block.AppVersion += "-wrong" },
			"wrong Block.Header.AppVersion",
		},
		{
			"ChainID wrong",
			func(block *types.Block) { block.ChainID = "not-the-real-one" },
			"wrong Block.Header.ChainID",
		},
		{
			"Height wrong",
			func(block *types.Block) { block.Height += 10 },
			"",
		},
		{
			"Time wrong",
			func(block *types.Block) { block.Time = block.Time.Add(-time.Second * 1) },
			"",
		},
		{
			"NumTxs wrong",
			func(block *types.Block) { block.NumTxs += 10 },
			"wrong Header.NumTxs",
		},
		{
			"TotalTxs wrong",
			func(block *types.Block) { block.TotalTxs += 10 },
			"wrong Block.Header.TotalTxs",
		},
		{
			"LastBlockID wrong",
			func(block *types.Block) { block.LastBlockID.PartsHeader.Total += 10 },
			"wrong Block.Header.LastBlockID",
		},
		{
			"LastCommitHash wrong",
			func(block *types.Block) { block.LastCommitHash = wrongHash },
			"wrong Header.LastCommitHash",
		},
		{
			"DataHash wrong",
			func(block *types.Block) { block.DataHash = wrongHash },
			"wrong Header.DataHash",
		},
		{
			"ValidatorsHash wrong",
			func(block *types.Block) { block.ValidatorsHash = wrongHash },
			"wrong Block.Header.ValidatorsHash",
		},
		{
			"NextValidatorsHash wrong",
			func(block *types.Block) { block.NextValidatorsHash = wrongHash },
			"wrong Block.Header.NextValidatorsHash",
		},
		{
			"ConsensusHash wrong",
			func(block *types.Block) { block.ConsensusHash = wrongHash },
			"wrong Block.Header.ConsensusHash",
		},
		{
			"AppHash mismatch",
			func(block *types.Block) { block.AppHash = wrongHash },
			fmt.Sprintf("wrong Block.Header.AppHash.  Expected %X, got %X", validHash, wrongHash),
		},
		{
			"LastResultsHash wrong",
			func(block *types.Block) { block.LastResultsHash = wrongHash },
			fmt.Sprintf("wrong Block.Header.LastResultsHash.  Expected %X, got %X", validHash, wrongHash),
		},
		{
			"Proposer wrong",
			func(block *types.Block) { block.ProposerAddress = wrongAddress },
			fmt.Sprintf("Block.Header.ProposerAddress, %X, is not a validator", wrongAddress),
		},
		{
			"Proposer invalid",
			func(block *types.Block) { block.ProposerAddress = invalidAddress /* zero */ },
			fmt.Sprintf("Block.Header.ProposerAddress, %X, is not a validator", invalidAddress),
		},
	}

	// Build up state for multiple heights
	for height := int64(1); height < validationTestsStopHeight; height++ {
		proposerAddr := state.Validators.GetProposer().Address
		state.AppHash = validHash
		state.LastResultsHash = validHash

		/*
		   Invalid blocks don't pass
		*/
		for _, tc := range testCases {
			block, _ := state.MakeBlock(height, makeTxs(height), lastCommit, proposerAddr)
			tc.malleateBlock(block)
			err := state.ValidateBlock(block)
			assert.ErrorContains(t, err, tc.expectedError, tc.name)
		}

		/*
		   A good block passes
		*/
		var err error
		state, _, lastCommit, err = makeAndCommitGoodBlock(state, height, lastCommit, proposerAddr, blockExec, privVals)
		require.NoError(t, err, "height %d", height)
	}
}

func TestValidateBlockCommit(t *testing.T) {
	t.Parallel()

	proxyApp := newTestApp()
	require.NoError(t, proxyApp.Start())
	defer proxyApp.Stop()

	state, stateDB, privVals := makeState(1, 1)
	blockExec := sm.NewBlockExecutor(stateDB, log.NewTestingLogger(t), proxyApp.Consensus(), mock.Mempool{})
	lastCommit := types.NewCommit(types.BlockID{}, nil)
	wrongPrecommitsCommit := types.NewCommit(types.BlockID{}, nil)
	badPrivVal := types.NewMockPV()

	for height := int64(1); height < validationTestsStopHeight; height++ {
		proposerAddr := state.Validators.GetProposer().Address
		if height > 1 {
			/*
				#2589: ensure state.LastValidators.VerifyCommit fails here
			*/
			// should be height-1 instead of height
			wrongHeightVote, err := types.MakeVote(height, state.LastBlockID, state.Validators, privVals[proposerAddr.String()], chainID)
			require.NoError(t, err, "height %d", height)
			wrongHeightCommit := types.NewCommit(state.LastBlockID, []*types.CommitSig{wrongHeightVote.CommitSig()})
			block, _ := state.MakeBlock(height, makeTxs(height), wrongHeightCommit, proposerAddr)
			err = state.ValidateBlock(block)
			_, isErrInvalidCommitHeight := err.(types.InvalidCommitHeightError)
			require.True(t, isErrInvalidCommitHeight, "expected InvalidCommitHeightError at height %d but got: %v", height, err)

			/*
				#2589: test len(block.LastCommit.Precommits) == state.LastValidators.Size()
			*/
			block, _ = state.MakeBlock(height, makeTxs(height), wrongPrecommitsCommit, proposerAddr)
			err = state.ValidateBlock(block)
			_, isErrInvalidCommitPrecommits := err.(types.InvalidCommitPrecommitsError)
			require.True(t, isErrInvalidCommitPrecommits, "expected InvalidCommitPrecommitsError at height %d but got: %v", height, err)
		}

		/*
			A good block passes
		*/
		var err error
		var blockID types.BlockID
		state, blockID, lastCommit, err = makeAndCommitGoodBlock(state, height, lastCommit, proposerAddr, blockExec, privVals)
		require.NoError(t, err, "height %d", height)

		/*
			wrongPrecommitsCommit is fine except for the extra bad precommit
		*/
		goodVote, err := types.MakeVote(height, blockID, state.Validators, privVals[proposerAddr.String()], chainID)
		require.NoError(t, err, "height %d", height)
		badVote := &types.Vote{
			ValidatorAddress: badPrivVal.PubKey().Address(),
			ValidatorIndex:   0,
			Height:           height,
			Round:            0,
			Timestamp:        tmtime.Now(),
			Type:             types.PrecommitType,
			BlockID:          blockID,
		}
		err = badPrivVal.SignVote(chainID, goodVote)
		require.NoError(t, err, "height %d", height)
		wrongPrecommitsCommit = types.NewCommit(blockID, []*types.CommitSig{goodVote.CommitSig(), badVote.CommitSig()})
	}
}
