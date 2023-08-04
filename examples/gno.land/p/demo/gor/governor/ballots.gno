package governor

import (
	"errors"
	"gno.land/p/demo/avl"
	std "std"
)

// ---------------------------------------------------
type Votes struct {
	vs *avl.Tree
}

func NewVotes() *Votes {
	return &Votes{
		vs: avl.NewTree(),
	}
}

func (vs *Votes) setVote(account std.Address, vt VoteType) {
	vs.vs.Set(account.String(), vt)
}

type ProposalVotes struct {
	propVotes *avl.Tree // proposalId => Votes
}

func NewProposalVotes() *ProposalVotes {
	return &ProposalVotes{
		propVotes: avl.NewTree(),
	}
}

var _ Tally = (*ProposalVotes)(nil)

func (pv *ProposalVotes) hasVoted(proposalId string, account std.Address) bool {
	vsI, ok := pv.propVotes.Get(proposalId)
	if !ok {
		panic("GCS: proposal not exist")
	}
	votes := vsI.(*Votes)

	_, ok = votes.vs.Get(account.String())
	if !ok {
		return false
	}
	return true
}

// this method counts who votes, in a map vote => voteType
// 1:1, voting power will be calculated in Tally process
func (pv *ProposalVotes) CastVote(propID string, account std.Address, vt VoteType) error {
	var votes *Votes

	vsI, ok := pv.propVotes.Get(propID)
	if !ok { // if vote not exist, means no one vote yet
		votes = NewVotes()
		votes.setVote(account, vt)
	} else { // has vote set for this proposal
		votes := vsI.(*Votes)
		_, ok = votes.vs.Get(account.String())
		// only vote when not voted before
		if !ok {
			// set vote state
			votes.setVote(account, vt)
		} else {
			return errors.New("GCS: You have voted for this before")
		}
	}
	// update
	pv.propVotes.Set(propID, votes)
	return nil
}

func (pv *ProposalVotes) Tally(propID string) (*TallyResult, error) {
	// TODO: or get from args?
	//denom := gcs.stake.GetDenom()
	//snapshot := std.GetHeight()

	// init voting power, to check quorum and veto
	var votingPower uint64 // total voting power, to determine if quorum or veto reached
	var result = NewTallyResult()

	// get all votes, and iterate to tally
	vsI, ok := pv.propVotes.Get(propID)
	if !ok {
		return new(TallyResult), errors.New("no votes for proposal")
	}
	// addr => voteType: for, against...
	votes := vsI.(*Votes)

	// iterate vote map, key is account, value is voteType
	votes.vs.Iterate("", "", func(key string, value interface{}) bool {
		// TODO: voting power is determined by different criteria, such like 1:1, staking based, etc
		//vp := governor.GetVotingPowerByAccount(key)
		votingPower += 1 // 1 is the default voting power
		// TODO: get voting power by setting, 1 member 1 vote, or calculated by different criterion
		vt := value.(VoteType)
		//result.tr[vt] += 1 // accumulate by vote type
		var count uint64
		icount, ok := result.tr.Get(vt.String())
		if ok {
			count = icount.(uint64) // else 0
		}
		count += 1
		result.tr.Set(vt.String(), count) // update

		return false
	})

	// now we have a voting power, and results
	println("accumulated voting power of all : ", votingPower)

	// pass when super majority achieved
	return result, nil
}
