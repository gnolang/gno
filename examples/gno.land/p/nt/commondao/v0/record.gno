package commondao

import (
	"gno.land/p/moul/addrset"
	"gno.land/p/nt/bptree/v0"
)

type (
	// VoteIterFn defines a callback to iterate votes.
	VoteIterFn func(Vote) (stop bool)

	// VotesCountIterFn defines a callback to iterate voted choices.
	VotesCountIterFn func(_ VoteChoice, voteCount int) (stop bool)

	// Vote defines a single vote. Its fields are unexported so instances
	// cannot be forged or reshaped outside the package: votes enter a
	// record only through CommonDAO.Vote's gates.
	Vote struct {
		addr   address
		choice VoteChoice
		reason string
	}
)

// NewVote creates a vote, validating the address and choice. It exists
// so external code can build records to independently re-verify a tally
// with TallyDefault; votes reach a DAO's own records only through
// CommonDAO.Vote.
func NewVote(addr address, choice VoteChoice, reason string) (Vote, error) {
	if !addr.IsValid() {
		return Vote{}, ErrInvalidVoterAddress
	}

	if choice != ChoiceYes && choice != ChoiceNo && choice != ChoiceAbstain {
		return Vote{}, ErrInvalidVoteChoice
	}

	return Vote{addr: addr, choice: choice, reason: reason}, nil
}

// Address returns the address of the account that submitted the vote.
func (v Vote) Address() address {
	return v.addr
}

// Choice returns the voted choice.
func (v Vote) Choice() VoteChoice {
	return v.choice
}

// Reason returns the optional reason for the vote.
func (v Vote) Reason() string {
	return v.reason
}

// ReadonlyVotingRecord defines a read only voting record. The copy
// shares tree nodes with the live record, so a held value can go stale
// in surprising ways: fetch it, read it, and re-fetch rather than
// holding it across votes.
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

// IterateVotesCount iterates voted choices with the amount of votes submitted for each.
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
	v := r.votes.Get(vote.addr.String())

	// When a previous vote exists update counter for the previous choice
	updated = r.votes.Set(vote.addr.String(), vote)
	if updated {
		prev := v.(Vote)
		r.count.Set(string(prev.choice), r.VoteCount(prev.choice)-1)
	}

	r.count.Set(string(vote.choice), r.VoteCount(vote.choice)+1)
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
		if !electorate.Has(v.addr) {
			return false
		}

		switch v.choice {
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
