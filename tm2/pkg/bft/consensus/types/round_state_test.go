package cstypes

import (
	"testing"

	"github.com/gnolang/gno/tm2/ordering"

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
	for i := 0; i < nval; i++ {
		precommits[i] = (&types.Vote{
			ValidatorAddress: crypto.AddressFromBytes(random.RandBytes(20)),
			Timestamp:        tmtime.Now(),
			BlockID:          blockID,
			Signature:        sig,
		}).CommitSig()
	}
	txs := make([]types.Tx, ntxs)
	for i := 0; i < ntxs; i++ {
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

func TestCompare(t *testing.T) {
	tests := []struct {
		name   string
		hrs1   HRS
		hrs2   HRS
		result ordering.Ordering
	}{
		{"Equal HRS", HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, ordering.Equal},
		{"HRS1 Lesser Height", HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, HRS{Height: 2, Round: 2, Step: RoundStepNewHeight}, ordering.Less},
		{"HRS1 Greater Height", HRS{Height: 2, Round: 2, Step: RoundStepNewHeight}, HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, ordering.Greater},
		{"Equal Height, HRS1 Lesser Round", HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, HRS{Height: 1, Round: 3, Step: RoundStepNewHeight}, ordering.Less},
		{"Equal Height, HRS1 Greater Round", HRS{Height: 1, Round: 3, Step: RoundStepNewHeight}, HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, ordering.Greater},
		{"Equal Height, Equal Round, HRS1 Lesser Step", HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, HRS{Height: 1, Round: 2, Step: RoundStepPropose}, ordering.Less},
		{"Equal Height, Equal Round, HRS1 Greater Step", HRS{Height: 1, Round: 2, Step: RoundStepPropose}, HRS{Height: 1, Round: 2, Step: RoundStepNewHeight}, ordering.Greater},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.hrs1.Compare(tt.hrs2)
			if result != tt.result {
				t.Errorf("Expected %d, got %d", tt.result, result)
			}
		})
	}

	// Test cases for all RoundStepType values
	t.Run("All RoundStepType Values", func(t *testing.T) {
		for step1 := RoundStepInvalid; step1 <= RoundStepCommit; step1++ {
			for step2 := RoundStepInvalid; step2 <= RoundStepCommit; step2++ {
				hrs1 := HRS{Height: 1, Round: 2, Step: step1}
				hrs2 := HRS{Height: 1, Round: 2, Step: step2}
				result := hrs1.Compare(hrs2)
				if step1 < step2 && result != ordering.Less {
					t.Errorf("Expected -1, got %d for %s < %s", result, step1, step2)
				} else if step1 > step2 && result != ordering.Greater {
					t.Errorf("Expected 1, got %d for %s > %s", result, step1, step2)
				} else if step1 == step2 && result != ordering.Equal {
					t.Errorf("Expected 0, got %d for %s == %s", result, step1, step2)
				}
			}
		}
	})
}
