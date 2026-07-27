package commondao

import (
	"errors"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/bptree/v0"
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
	votes bptree.BPTree // string(address) -> Vote
	count bptree.BPTree // string(choice) -> int
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
	if v := r.count.Get(string(c)); v != nil {
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
	if v := r.votes.Get(user.String()); v != nil {
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
	v := r.votes.Get(vote.Address.String())

	// When a previous vote exists update counter for the previous choice
	updated = r.votes.Set(vote.Address.String(), vote)
	if updated {
		prev := v.(Vote)
		r.count.Set(string(prev.Choice), r.VoteCount(prev.Choice)-1)
	}

	r.count.Set(string(vote.Choice), r.VoteCount(vote.Choice)+1)
	return
}

// TallyDefault applies the constitution's default Council voting rules
// over a proposal's electorate.
//
// Only votes cast by electorate members are counted. The tally denominator
// D is the electorate size minus the number of ABSTAIN votes: abstaining
// shrinks the denominator (deference), while not voting counts against
// passage (silence is opposition). With integer math:
//
//	D = |electorate| - abstains
//	supermajority:   passed    ⇔ D > 0 && 3*yes >= 2*D
//	simple majority: passed    ⇔ D > 0 && 2*yes > D
//	both:            dismissed ⇔ 2*no > D
//
// Passing is checked before dismissal; within one electorate both can never
// hold at once (yes+no <= D makes each pair contradictory). When D is zero
// or negative (an empty electorate, or every member abstained) the outcome
// stays pending: nothing can pass with zero YES votes.
func TallyDefault(r ReadonlyVotingRecord, electorate *addrset.ReadonlySet, t Threshold) Outcome {
	var yes, no, abstain int
	r.Iterate(0, r.Size(), false, func(v Vote) bool {
		if !electorate.Has(v.Address) {
			return false
		}

		switch v.Choice {
		case ChoiceYes:
			yes++
		case ChoiceNo:
			no++
		case ChoiceAbstain:
			abstain++
		}
		return false
	})

	d := electorate.Size() - abstain
	if d <= 0 {
		return OutcomePending
	}

	switch t {
	case ThresholdSimpleMajority:
		if 2*yes > d {
			return OutcomePassed
		}
	default:
		if 3*yes >= 2*d {
			return OutcomePassed
		}
	}

	if 2*no > d {
		return OutcomeDismissed
	}
	return OutcomePending
}
