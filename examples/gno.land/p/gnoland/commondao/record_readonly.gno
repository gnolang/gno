package commondao

import (
	"std"

	"gno.land/p/demo/avl"
)

// ReadOnlyVotingRecord defines an read only voting record.
type ReadOnlyVotingRecord struct {
	votes avl.Tree // string(address) -> Vote
	count avl.Tree // string(choice) -> int
}

// Size returns the total number of votes that record contains.
func (r ReadOnlyVotingRecord) Size() int {
	return r.votes.Size()
}

// Choices returns the voting choices that has been voted.
func (r ReadOnlyVotingRecord) Choices() []VoteChoice {
	var choices []VoteChoice
	r.count.Iterate("", "", func(k string, v any) bool {
		choices = append(choices, VoteChoice(k))
		return false
	})
	return choices
}

// Iterate iterates voting record votes.
func (r ReadOnlyVotingRecord) Iterate(fn VoteIterFn) bool {
	return r.votes.Iterate("", "", func(_ string, v any) bool {
		return fn(v.(Vote))
	})
}

// VoteCount returns the number of votes for a single voting choice.
func (r ReadOnlyVotingRecord) VoteCount(c VoteChoice) int {
	if v, found := r.count.Get(string(c)); found {
		return v.(int)
	}
	return 0
}

// HasVoted checks if an account already voted.
func (r ReadOnlyVotingRecord) HasVoted(user std.Address) bool {
	return r.votes.Has(user.String())
}
