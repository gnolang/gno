package commondao

import (
	"errors"
	"math"

	"gno.land/p/nt/avl"
)

// ErrVoteExists indicates that a user already voted.
var ErrVoteExists = errors.New("user already voted")

type (
	// VoteIterFn defines a callback to iterate votes.
	VoteIterFn func(Vote) (stop bool)

	// VotesCountIterFn defines a callback to iterate voted choices.
	VotesCountIterFn func(_ VoteChoice, voteCount int) (stop bool)

	// Vote defines a single vote.
	Vote struct {
		// Address is the address of the user that this vote belons to.
		Address address

		// Choice contains the voted choice.
		Choice VoteChoice

		// Reason contains an optional reason for the vote.
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

// Iterate iterates voting record votes.
func (r ReadonlyVotingRecord) Iterate(offset, count int, reverse bool, fn VoteIterFn) bool {
	cb := func(_ string, v any) bool { return fn(v.(Vote)) }
	if reverse {
		return r.votes.ReverseIterateByOffset(offset, count, cb)
	}
	return r.votes.IterateByOffset(offset, count, cb)
}

// IterateVotesCount iterates voted choices with the amount of votes submited for each.
func (r ReadonlyVotingRecord) IterateVotesCount(fn VotesCountIterFn) bool {
	return r.count.Iterate("", "", func(k string, v any) bool {
		return fn(VoteChoice(k), v.(int))
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
func (r ReadonlyVotingRecord) HasVoted(user address) bool {
	return r.votes.Has(user.String())
}

// GetVote returns a vote.
func (r ReadonlyVotingRecord) GetVote(user address) (_ Vote, found bool) {
	if v, found := r.votes.Get(user.String()); found {
		return v.(Vote), true
	}
	return Vote{}, false
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

	r.IterateVotesCount(func(c VoteChoice, count int) bool {
		if currentCount <= count {
			choice = c
			prevCount = currentCount
			currentCount = count
		}
		return false
	})

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

	r.IterateVotesCount(func(c VoteChoice, count int) bool {
		// Don't consider explicit abstentions or invalid votes
		if c == ChoiceAbstain || c == ChoiceNone {
			return false
		}

		if currentCount < count {
			choice = c
			currentCount = count
			isMajority = true
		} else if currentCount == count {
			isMajority = false
		}
		return false
	})

	if isMajority {
		return choice, true
	}
	return ChoiceNone, false
}

// CollectVotes returns an voting record containing votes of members from one or more groups.
// Returned tree uses member account address as key and `commondao.Vote` as value.
func CollectVotes(ctx VotingContext, groups ...string) (*VotingRecord, error) {
	if len(groups) == 0 {
		return nil, errors.New("one or more group names are required to collect votes")
	}

	var (
		votes    VotingRecord
		grouping = ctx.Members.Grouping()
	)

	for _, name := range groups {
		group, found := grouping.Get(name)
		if !found {
			return nil, errors.New("member group not found: " + name)
		}

		group.Members().IterateByOffset(0, group.Members().Size(), func(member address) bool {
			v, found := ctx.VotingRecord.GetVote(member)
			if found {
				votes.AddVote(v)
			}
			return false
		})
	}
	return &votes, nil
}
