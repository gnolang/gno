package governor

import (
	"errors"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/maths"
	"std"
)

type VoteType int

const (
	Against VoteType = iota
	For
	Abstain
	NoWithVeto
)

func (vt VoteType) String() string {
	switch vt {
	case Against:
		return "against"
	case For:
		return "for"
	case Abstain:
		return "abstain"
	case NoWithVeto:
		return "veto"
	default:
		panic("no such VoteType")
	}
}

func str2VoteType(str string) VoteType {
	switch str {
	case "for":
		return For
	case "against":
		return Against
	case "abstain":
		return Abstain
	case "veto":
		return NoWithVeto
	default:
		panic("no such vote type")
	}
}

type Ballots interface {
	Tally(propID string) (*TallyResult, error)
	CastVote(propID string, account std.Address, vp VoteType, weight int32, memo string) error
}

type TallyResult struct {
	tr          *avl.Tree // account -> voteType
	votingPower int32
}

func NewTallyResult() *TallyResult {
	return &TallyResult{
		tr: avl.NewTree(),
	}
}

func (trt *TallyResult) GetVotesByType(vt VoteType) int32 {
	count, ok := trt.tr.Get(vt.String())
	if ok {
		return count.(int32)
	}
	return 0
}

type Vote struct {
	voter  std.Address
	vt     VoteType
	weight int32
	memo   string
}

func NewVote(voter std.Address, vt VoteType, weight int32, memo string) *Vote {
	return &Vote{
		voter:  voter,
		vt:     vt,
		weight: weight,
		memo:   memo,
	}
}

// ---------------------------------------------------
type Votes struct {
	vs *avl.Tree
}

func NewVotes() *Votes {
	return &Votes{
		vs: avl.NewTree(),
	}
}

func (vs *Votes) setVote(account std.Address, vote *Vote) {
	vs.vs.Set(account.String(), vote)
}

type ballots struct {
	propVotes *avl.Tree // proposalId => Votes
}

func NewBallots() *ballots {
	return &ballots{
		propVotes: avl.NewTree(),
	}
}

var _ Ballots = (*ballots)(nil)

func (pv *ballots) hasVoted(proposalId string, account std.Address) bool {
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
// 1:1, voting power will be calculated in Ballots process
func (pv *ballots) CastVote(propID string, voter std.Address, vt VoteType, weight int32, memo string) error {
	vote := NewVote(voter, vt, weight, memo)
	var votes *Votes
	vsI, ok := pv.propVotes.Get(propID)
	if !ok { // if vote not exist, means no one vote yet
		votes = NewVotes()
		votes.setVote(voter, vote)
	} else { // has vote set for this proposal
		votes = vsI.(*Votes)
		_, ok = votes.vs.Get(voter.String())
		// only vote when not voted before
		if !ok {
			// set vote state
			votes.setVote(voter, vote)
		} else {
			return errors.New("You have voted for this before")
		}
	}
	// update
	pv.propVotes.Set(propID, votes)
	return nil
}

func (pv *ballots) Tally(propID string) (*TallyResult, error) {
	println("tally at height of: ", std.GetHeight())
	// TODO: or get from args?
	//denom := gcs.stake.GetDenom()
	//snapshot := std.GetHeight()

	// init voting power, to check quorum and veto
	var votingPower int32 // total voting power, to determine if quorum or veto reached
	var result = NewTallyResult()

	// get all votes, and iterate to ballots
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

		vote, ok := value.(*Vote)
		if !ok {
			panic("should not happen")
		}

		votingPower += vote.weight // 1 is the default voting power
		// TODO: get voting power by setting, 1 member 1 vote, or calculated by different criterion
		// set result by vote type
		vt := vote.vt
		var count int32
		icount, ok := result.tr.Get(vt.String())
		if ok {
			count = icount.(int32) // else 0
		}
		count += vote.weight
		result.tr.Set(vt.String(), count) // update
		result.votingPower = votingPower

		return false
	})

	// now we have a voting power, and results
	println("accumulated voting power of all : ", votingPower)

	// pass when super majority achieved
	return result, nil
}

// ---------------------------------------------------------------
// check if quorum reached
// TODO: should voting power obtained at height?
func (gvr *Governor) quorumReached(totalVotingPower int32, height int64) bool {
	println("quorum reached check, ", totalVotingPower, gvr.TotalVotingPowerAt(height))
	left := maths.NewRat(totalVotingPower, gvr.TotalVotingPower())
	right := gvr.gs.getQuorum()
	r := left.Cmp(right)
	if r == -1 {
		return false
	}
	return true
}

// veto reached, 33.4
func (gvr *Governor) vetoReached(result *TallyResult) bool {
	// all votes
	votes := result.GetVotesByType(For) + result.GetVotesByType(Abstain) + result.GetVotesByType(Against) + result.GetVotesByType(NoWithVeto)
	left := maths.NewRat(result.GetVotesByType(NoWithVeto), votes)
	right := gvr.gs.getVetoThreshold()
	r := left.Cmp(right)
	if r == -1 {
		return false
	}
	return true
}

// TODO: super majority required
func (gvr *Governor) voteSucceeded(result *TallyResult) bool {
	// get all votes count
	votes := result.GetVotesByType(For) + result.GetVotesByType(Abstain) + result.GetVotesByType(Against) + result.GetVotesByType(NoWithVeto)
	left := maths.NewRat(result.GetVotesByType(For), votes)
	right := gvr.gs.getSucceedThreshold()

	r := left.Cmp(right)
	if r == -1 {
		return false
	}
	return true
}

// TODO: undetermined rules
func (gvr *Governor) isBurnDeposit(proposalId string) bool {
	return false
}
