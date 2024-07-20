package simpledao

import (
	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/membstore"
	"gno.land/p/gov/dao"
)

var ErrAlreadyVoted = errors.New("vote already cast")

// votes is a simple weighted voting system
type votes struct {
	// tally cache to keep track of active
	// yes / no / abstain votes
	yays     uint64
	nays     uint64
	abstains uint64

	voters *avl.Tree // std.Address -> dao.VoteOption
}

// newVotes creates a new weighted voting system instance
func newVotes() *votes {
	return &votes{
		voters: avl.NewTree(),
	}
}

// castVote casts a single vote in the name of the given member
func (v *votes) castVote(member membstore.Member, option dao.VoteOption) error {
	// Check if the member voted already
	address := member.Address.String()

	_, voted := v.voters.Get(address)
	if voted {
		return ErrAlreadyVoted
	}

	// Update the tally
	switch option {
	case dao.YesVote:
		v.yays += member.VotingPower
	case dao.AbstainVote:
		v.abstains += member.VotingPower
	default:
		v.nays += member.VotingPower
	}

	// Save the voting status
	v.voters.Set(address, option)

	return nil
}
