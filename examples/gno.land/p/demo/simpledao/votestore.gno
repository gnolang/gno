package simpledao

import (
	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/dao"
	"gno.land/p/demo/membstore"
)

var ErrAlreadyVoted = errors.New("vote already cast")

// tally is a simple vote tally system
type tally struct {
	// tally cache to keep track of active
	// yes / no / abstain votes
	yays     uint64
	nays     uint64
	abstains uint64

	voters *avl.Tree // std.Address -> dao.VoteOption
}

// newTally creates a new tally system instance
func newTally() *tally {
	return &tally{
		voters: avl.NewTree(),
	}
}

// castVote casts a single vote in the name of the given member
func (t *tally) castVote(member membstore.Member, option dao.VoteOption) error {
	// Check if the member voted already
	address := member.Address.String()

	_, voted := t.voters.Get(address)
	if voted {
		return ErrAlreadyVoted
	}

	// Update the tally
	switch option {
	case dao.YesVote:
		t.yays += member.VotingPower
	case dao.AbstainVote:
		t.abstains += member.VotingPower
	default:
		t.nays += member.VotingPower
	}

	// Save the voting status
	t.voters.Set(address, option)

	return nil
}
