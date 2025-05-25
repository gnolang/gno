package types

import (
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
)

func MakeCommit(blockID BlockID, height int64, round int,
	voteSet *VoteSet, validators []PrivValidator,
) (*Commit, error) {
	// all sign
	for i := range validators {
		vote := &Vote{
			ValidatorAddress: validators[i].PubKey().Address(),
			ValidatorIndex:   i,
			Height:           height,
			Round:            round,
			Type:             PrecommitType,
			BlockID:          blockID,
			Timestamp:        tmtime.Now(),
		}

		if _, err := signAddVote(validators[i], vote, voteSet); err != nil {
			return nil, err
		}
	}

	return voteSet.MakeCommit(), nil
}

func signAddVote(privVal PrivValidator, vote *Vote, voteSet *VoteSet) (signed bool, err error) {
	err = privVal.SignVote(voteSet.ChainID(), vote)
	if err != nil {
		return false, err
	}
	return voteSet.AddVote(vote)
}

func MakeVote(height int64, blockID BlockID, valSet *ValidatorSet, privVal PrivValidator, chainID string) (*Vote, error) {
	addr := privVal.PubKey().Address()
	idx, _ := valSet.GetByAddress(addr)
	vote := &Vote{
		ValidatorAddress: addr,
		ValidatorIndex:   idx,
		Height:           height,
		Round:            0,
		Timestamp:        tmtime.Now(),
		Type:             PrecommitType,
		BlockID:          blockID,
	}
	if err := privVal.SignVote(chainID, vote); err != nil {
		return nil, err
	}
	return vote, nil
}
