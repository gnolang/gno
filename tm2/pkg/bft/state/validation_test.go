package state_test

import (
	"testing"
	"time"

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

	// some bad values
	wrongHash := tmhash.Sum([]byte("this hash is wrong"))

	// Manipulation of any header field causes failure.
	testCases := []struct {
		name          string
		malleateBlock func(block *types.Block)
	}{
		{"BlockVersion wrong", func(block *types.Block) { block.Version += "-wrong" }},
		{"AppVersion wrong", func(block *types.Block) { block.AppVersion += "-wrong" }},
		{"ChainID wrong", func(block *types.Block) { block.ChainID = "not-the-real-one" }},
		{"Height wrong", func(block *types.Block) { block.Height += 10 }},
		{"Time wrong", func(block *types.Block) { block.Time = block.Time.Add(-time.Second * 1) }},
		{"NumTxs wrong", func(block *types.Block) { block.NumTxs += 10 }},
		{"TotalTxs wrong", func(block *types.Block) { block.TotalTxs += 10 }},

		{"LastBlockID wrong", func(block *types.Block) { block.LastBlockID.PartsHeader.Total += 10 }},
		{"LastCommitHash wrong", func(block *types.Block) { block.LastCommitHash = wrongHash }},
		{"DataHash wrong", func(block *types.Block) { block.DataHash = wrongHash }},

		{"ValidatorsHash wrong", func(block *types.Block) { block.ValidatorsHash = wrongHash }},
		{"NextValidatorsHash wrong", func(block *types.Block) { block.NextValidatorsHash = wrongHash }},
		{"ConsensusHash wrong", func(block *types.Block) { block.ConsensusHash = wrongHash }},
		{"AppHash wrong", func(block *types.Block) { block.AppHash = wrongHash }},
		{"LastResultsHash wrong", func(block *types.Block) { block.LastResultsHash = wrongHash }},

		{"Proposer wrong", func(block *types.Block) { block.ProposerAddress = ed25519.GenPrivKey().PubKey().Address() }},
		{"Proposer invalid", func(block *types.Block) { block.ProposerAddress = crypto.Address{} /* zero */ }},
	}

	// Build up state for multiple heights
	for height := int64(1); height < validationTestsStopHeight; height++ {
		proposerAddr := state.Validators.GetProposer().Address
		/*
			Invalid blocks don't pass
		*/
		for _, tc := range testCases {
			block, _ := state.MakeBlock(height, makeTxs(height), lastCommit, proposerAddr)
			tc.malleateBlock(block)
			err := blockExec.ValidateBlock(state, block)
			require.Error(t, err, tc.name)
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
			err = blockExec.ValidateBlock(state, block)
			_, isErrInvalidCommitHeight := err.(types.InvalidCommitHeightError)
			require.True(t, isErrInvalidCommitHeight, "expected InvalidCommitHeightError at height %d but got: %v", height, err)

			/*
				#2589: test len(block.LastCommit.Precommits) == state.LastValidators.Size()
			*/
			block, _ = state.MakeBlock(height, makeTxs(height), wrongPrecommitsCommit, proposerAddr)
			err = blockExec.ValidateBlock(state, block)
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
			ValidatorAddress: badPrivVal.GetPubKey().Address(),
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
