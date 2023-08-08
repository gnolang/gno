package governor

import "gno.land/p/demo/maths"

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

// ---------------------------------------------
// check if quorum reached
func (gvr *Governor) quorumReached(totalVotingPower int32) bool {
	left := maths.NewRat(totalVotingPower, gvr.votingPower.Total())
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
