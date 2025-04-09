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
	// VoteIterFn defines a callback to iterate votes.
	VoteIterFn func(Vote) (stop bool)

	// Vote defines a single vote.
	Vote struct {
		// Address is the address of the user that this vote belons to.
		Address std.Address

		// Choice contains the voted choice.
		Choice VoteChoice

		// Reason contains the reason for the vote.
		Reason string

		// Context can store any custom voting values related to the vote.
		//
		// Warning: When using context be careful if references/pointers are
		// assigned to it because they could potentially be accessed anywhere,
		// which could lead to unwanted indirect modifications.
		Context any
	}
)

// ReadonlyVotingRecord defines an read only voting record.
type ReadonlyVotingRecord struct {
	votes avl.Tree // string(address) -> Vote
	count avl.Tree // string(choice) -> int
}

// Size returns the total number of votes that record contains.
func (r ReadonlyVotingRecord) Size() int {
	return r.votes.Size()
}

// Choices returns the voting choices that has been voted.
func (r ReadonlyVotingRecord) Choices() []VoteChoice {
	choices := make([]VoteChoice, 0, r.count.Size())
	r.count.Iterate("", "", func(k string, v any) bool {
		choices = append(choices, VoteChoice(k))
		return false
	})
	return choices
}

// Iterate iterates voting record votes.
func (r ReadonlyVotingRecord) Iterate(offset, count int, fn VoteIterFn) bool {
	return r.votes.IterateByOffset(offset, count, func(_ string, v any) bool {
		return fn(v.(Vote))
	})
}

// VoteCount returns the number of votes for a single voting choice.
func (r ReadonlyVotingRecord) VoteCount(c VoteChoice) int {
	if v, found := r.count.Get(string(c)); found {
		return v.(int)
	}
	return 0
}

// HasVoted checks if an account already voted.
func (r ReadonlyVotingRecord) HasVoted(user std.Address) bool {
	return r.votes.Has(user.String())
}

// VotingRecord stores accounts that voted and vote choices.
type VotingRecord struct {
	ReadonlyVotingRecord
}

// Readonly returns a read only voting record.
func (r VotingRecord) Readonly() ReadonlyVotingRecord {
	return r.ReadonlyVotingRecord
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

// FindMostVotedChoice returns the most voted choice.
// ChoiceNone is returned when there is a tie between different
// voting choices or when the voting record has are no votes.
func FindMostVotedChoice(r ReadonlyVotingRecord) VoteChoice {
	var (
		choice                  VoteChoice
		currentCount, prevCount int
	)

	for _, c := range r.Choices() {
		count := r.VoteCount(c)
		if currentCount <= count {
			choice = c
			prevCount = currentCount
			currentCount = count
		}
	}

	if prevCount < currentCount {
		return choice
	}
	return ChoiceNone
}

// SelectChoiceByAbsoluteMajority select the vote choice by absolute majority.
// Vote choice is a majority when chosen by more than half of the votes.
// Absolute majority considers abstentions when counting votes.
func SelectChoiceByAbsoluteMajority(r ReadonlyVotingRecord, membersCount int) (VoteChoice, bool) {
	choice := FindMostVotedChoice(r)
	if choice != ChoiceNone && r.VoteCount(choice) > int(membersCount/2) {
		return choice, true
	}
	return ChoiceNone, false
}

// SelectChoiceBySuperMajority select the vote choice by super majority using a 2/3s threshold.
// Abstentions are considered when calculating the super majority choice.
func SelectChoiceBySuperMajority(r ReadonlyVotingRecord, membersCount int) (VoteChoice, bool) {
	if membersCount < 3 {
		return ChoiceNone, false
	}

	choice := FindMostVotedChoice(r)
	if choice != ChoiceNone && r.VoteCount(choice) >= int(math.Ceil((2*float64(membersCount))/3)) {
		return choice, true
	}
	return ChoiceNone, false
}

// SelectChoiceByPlurality selects the vote choice by plurality.
// The choice will be considered a majority if it has votes and if there is no other
// choice with the same number of votes. A tie won't be considered majority.
func SelectChoiceByPlurality(r ReadonlyVotingRecord) (VoteChoice, bool) {
	var (
		choice       VoteChoice
		currentCount int
		isMajority   bool
	)

	for _, c := range r.Choices() {
		if c == ChoiceAbstain || c == ChoiceNone {
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
	return ChoiceNone, false
}
