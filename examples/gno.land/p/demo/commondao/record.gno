package commondao

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
)

// ErrVoteExists indicates that a user already voted.
var ErrVoteExists = errors.New("user already voted")

type (
	// Vote defines a single vote.
	Vote struct {
		Address std.Address
		Choice  VoteChoice
	}

	// VotingRecord stores accounts that voted and vote choices.
	VotingRecord struct {
		votes avl.Tree // string(address) -> VoteChoice
		count avl.Tree // string(choice) -> int
	}
)

// Votes returns the list of all votes.
func (r VotingRecord) Votes() []Vote {
	var votes []Vote
	r.votes.Iterate("", "", func(k string, v interface{}) bool {
		votes = append(votes, Vote{
			Address: std.Address(k),
			Choice:  v.(VoteChoice),
		})
		return false
	})
	return votes
}

// VoteCount returns the number of votes for a single voting choice.
func (r VotingRecord) VoteCount(c VoteChoice) int {
	if v, found := r.count.Get(string(c)); found {
		return v.(int)
	}
	return 0
}

// HasVoted checks if an account already voted.
func (r VotingRecord) HasVoted(user std.Address) bool {
	return r.votes.Has(user.String())
}

// AddVote adds a vote.
// Users are allowd to vote only once.
func (r *VotingRecord) AddVote(user std.Address, c VoteChoice) error {
	if r.HasVoted(user) {
		return ErrVoteExists
	}

	r.votes.Set(user.String(), c)
	r.count.Set(string(c), r.VoteCount(c)+1)
	return nil
}

// GetProvableMajorityChoice returns the choice voted by the majority.
// The result is only valid if there is a majority.
// Caller must validate that the returned choice represents a majority.
func (r VotingRecord) GetProvableMajorityChoice() VoteChoice {
	var (
		choice       VoteChoice
		currentCount int
	)

	r.count.Iterate("", "", func(k string, v interface{}) bool {
		count := v.(int)
		if currentCount < count {
			choice = VoteChoice(k)
			currentCount = count
		}
		return false
	})
	return choice
}
