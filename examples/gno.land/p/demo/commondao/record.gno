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

	// ReadOnlyVotingRecord defines an interface for read only voting records.
	ReadOnlyVotingRecord interface {
		// VoteChoices returns the voting choices that has been voted.
		VoteChoices() []VoteChoice

		// Votes returns the list of all votes.
		Votes() []Vote

		// VoteCount returns the number of votes for a single voting choice.
		VoteCount(VoteChoice) int

		// HasVoted checks if an account already voted.
		HasVoted(std.Address) bool

		// GetProvableMajorityChoice returns the choice voted by the majority.
		// The result is only valid if there is a majority.
		// Caller must validate that the returned choice represents a majority.
		GetProvableMajorityChoice() VoteChoice
	}
)

// VotingRecord stores accounts that voted and vote choices.
type VotingRecord struct {
	votes avl.Tree // string(address) -> VoteChoice
	count avl.Tree // string(choice) -> int
}

// VoteChoices returns the voting choices that has been voted.
func (r VotingRecord) VoteChoices() []VoteChoice {
	var choices []VoteChoice
	r.count.Iterate("", "", func(k string, v interface{}) bool {
		choices = append(choices, VoteChoice(k))
		return false
	})
	return choices
}

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

// SelectChoiceByAbsoluteMajority select the vote choice by absolute majority.
// Vote choice is a majority when chosen by more than half of the votes.
// Absolute majority considers abstentions when counting votes.
func SelectChoiceByAbsoluteMajority(r ReadOnlyVotingRecord, memberCount int) (VoteChoice, bool) {
	choice := r.GetProvableMajorityChoice()
	return choice, r.VoteCount(choice) > int(memberCount/2)
}

// SelectChoiceBySuperMajority select the vote choice by super majority using a 2/3s threshold.
// Abstentions are not considered when calculating the super majority choice.
func SelectChoiceBySuperMajority(r ReadOnlyVotingRecord) (VoteChoice, bool) {
	var count int
	for _, v := range r.Votes() {
		if v.Choice != ChoiceAbstain {
			count++
		}
	}

	choice := r.GetProvableMajorityChoice()
	return choice, r.VoteCount(choice) > int((2*count)/3)
}

// SelectChoiceByPlurality selects the vote choice by plurality.
// The choice will be considered a majority if it has votes and if there is no other
// choice with the same number of votes. A tie won't be considered majority.
func SelectChoiceByPlurality(r ReadOnlyVotingRecord) (VoteChoice, bool) {
	var (
		choice       VoteChoice
		currentCount int
		isMajority   bool
	)

	for _, c := range r.VoteChoices() {
		if c == ChoiceAbstain {
			continue
		}

		count := r.VoteCount(c)
		if currentCount < count {
			choice = c
			currentCount = count
			isMajority = true
		} else if currentCount == count {
			isMajority = false
		}
	}
	return choice, isMajority
}
