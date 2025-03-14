package commondao

import (
	"errors"
	"math"
	"std"
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
		Context any
	}
)

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

// GetProvableMajorityChoice returns the choice voted by the majority.
// The result is only valid if there is a majority.
// Caller must validate that the returned choice represents a majority.
func GetProvableMajorityChoice(r ReadonlyVotingRecord) VoteChoice { // TODO: Rethink this function
	var (
		choice       VoteChoice
		currentCount int
	)

	r.count.Iterate("", "", func(c string, v any) bool {
		count := v.(int)
		if currentCount < count {
			choice = VoteChoice(c)
			currentCount = count
		}
		return false
	})
	return choice
}

// SelectChoiceByAbsoluteMajority select the vote choice by absolute majority.
// Vote choice is a majority when chosen by more than half of the votes.
// Absolute majority considers abstentions when counting votes.
func SelectChoiceByAbsoluteMajority(r ReadonlyVotingRecord, membersCount int) (VoteChoice, bool) {
	choice := GetProvableMajorityChoice(r)
	if r.VoteCount(choice) > int(membersCount/2) {
		return choice, true
	}
	return "", false
}

// SelectChoiceBySuperMajority select the vote choice by super majority using a 2/3s threshold.
// Abstentions are not considered when calculating the super majority choice.
func SelectChoiceBySuperMajority(r ReadonlyVotingRecord, membersCount int) (VoteChoice, bool) {
	if membersCount < 3 {
		return "", false
	}

	choice := GetProvableMajorityChoice(r)
	if r.VoteCount(choice) >= int(math.Ceil((2*float64(membersCount))/3)) {
		return choice, true
	}
	return "", false
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
	return "", false
}
