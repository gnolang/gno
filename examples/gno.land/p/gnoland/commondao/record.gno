package commondao

import (
	"errors"
	"math"
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
		Reason  string
	}

	// ReadOnlyVotingRecord defines an interface for read only voting records.
	ReadOnlyVotingRecord interface {
		// Size returns the total number of votes that record contains.
		Size() int

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
	votes avl.Tree // string(address) -> Vote
	count avl.Tree // string(choice) -> int
}

// Size returns the total number of votes that record contains.
func (r VotingRecord) Size() int {
	return r.votes.Size()
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
		votes = append(votes, v.(Vote))
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

// AddVote adds a vote to the voting record.
// If a vote for the same user already exists is overwritten.
func (r *VotingRecord) AddVote(vote Vote) (updated bool) {
	// Get previous member vote if it exists
	v, _ := r.votes.Get(vote.Address.String())

	// When a previous vote exists update counter for the previous choice
	updated = r.votes.Set(vote.Address.String(), vote)
	if updated {
		prev := v.(Vote)
		r.count.Set(string(prev.Choice), r.VoteCount(prev.Choice)-1)
	}

	r.count.Set(string(vote.Choice), r.VoteCount(vote.Choice)+1)
	return
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
func SelectChoiceByAbsoluteMajority(r ReadOnlyVotingRecord, membersCount int) (VoteChoice, bool) {
	choice := r.GetProvableMajorityChoice()
	if r.VoteCount(choice) > int(membersCount/2) {
		return choice, true
	}
	return "", false
}

// SelectChoiceBySuperMajority select the vote choice by super majority using a 2/3s threshold.
// Abstentions are not considered when calculating the super majority choice.
func SelectChoiceBySuperMajority(r ReadOnlyVotingRecord, membersCount int) (VoteChoice, bool) {
	if membersCount < 3 {
		return "", false
	}

	choice := r.GetProvableMajorityChoice()
	if r.VoteCount(choice) >= int(math.Ceil((2*float64(membersCount))/3)) {
		return choice, true
	}
	return "", false
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

	if isMajority {
		return choice, true
	}
	return "", false
}
