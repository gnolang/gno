package cstypes

import (
	"testing"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/random"
)

func BenchmarkRoundStateDeepCopy(b *testing.B) {
	b.StopTimer()

	// Random validators
	nval, ntxs := 100, 100
	vset, _ := types.RandValidatorSet(nval, 1)
	precommits := make([]*types.CommitSig, nval)
	blockID := types.BlockID{
		Hash: random.RandBytes(20),
		PartsHeader: types.PartSetHeader{
			Hash: random.RandBytes(20),
		},
	}
	sig := make([]byte, ed25519.SignatureSize)
	for i := range nval {
		precommits[i] = (&types.Vote{
			ValidatorAddress: crypto.AddressFromBytes(random.RandBytes(20)),
			Timestamp:        tmtime.Now(),
			BlockID:          blockID,
			Signature:        sig,
		}).CommitSig()
	}
	txs := make([]types.Tx, ntxs)
	for i := range ntxs {
		txs[i] = random.RandBytes(100)
	}
	// Random block
	block := &types.Block{
		Header: types.Header{
			ChainID:         random.RandStr(12),
			Time:            tmtime.Now(),
			LastBlockID:     blockID,
			LastCommitHash:  random.RandBytes(20),
			DataHash:        random.RandBytes(20),
			ValidatorsHash:  random.RandBytes(20),
			ConsensusHash:   random.RandBytes(20),
			AppHash:         random.RandBytes(20),
			LastResultsHash: random.RandBytes(20),
		},
		Data: types.Data{
			Txs: txs,
		},
		LastCommit: types.NewCommit(blockID, precommits),
	}
	parts := block.MakePartSet(4096)
	// Random Proposal
	proposal := &types.Proposal{
		Timestamp: tmtime.Now(),
		BlockID:   blockID,
		Signature: sig,
	}
	// Random HeightVoteSet
	// TODO: hvs :=

	rs := &RoundState{
		StartTime:          tmtime.Now(),
		CommitTime:         tmtime.Now(),
		Validators:         vset,
		Proposal:           proposal,
		ProposalBlock:      block,
		ProposalBlockParts: parts,
		LockedBlock:        block,
		LockedBlockParts:   parts,
		ValidBlock:         block,
		ValidBlockParts:    parts,
		Votes:              nil, // TODO
		LastCommit:         nil, // TODO
		LastValidators:     vset,
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		amino.DeepCopy(rs)
	}
}
