package simpledao

import (
	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/gov/dao"
)

var errAlreadyVoted = errors.New("vote already cast")

// votes is a simple weighted voting system
type votes struct {
	// tally cache to keep track of active for / against votes
	yays uint64
	nays uint64

	voters *avl.Tree // std.Address -> dao.VoteOption
}

// newVotes creates a new weighted voting system instance
func newVotes() *votes {
	return &votes{
		voters: avl.NewTree(),
	}
}

// castVote casts a single vote in the name of the given member
func (v *votes) castVote(member dao.Member, option dao.VoteOption) error {
	// Check if the member voted already
	address := member.Address.String()

	_, voted := v.voters.Get(address)
	if voted {
		return errAlreadyVoted
	}

	// Update the tally
	if option == dao.YesVote {
		v.yays += member.VotingPower
	} else {
		v.nays += member.VotingPower
	}

	// Save the voting status
	v.voters.Set(address, option)

	return nil
}

// getTally returns the yay, and nay count, respectively
func (v *votes) getTally() (uint64, uint64) {
	return v.yays, v.nays
}

// getVoters returns the current voter set
func (v *votes) getVoters() *avl.Tree {
	return v.voters
}
