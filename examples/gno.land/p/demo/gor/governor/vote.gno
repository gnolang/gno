package governor

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

// ---------------------------------------------
// check if quorum reached
func (gvr *Governor) quorumReached(totalVotingPower uint64) bool {
	return gvr.gs.quorum(gvr.votingPower.Total()) <= totalVotingPower
}

// veto reached, 33.4
func (gvr *Governor) vetoReached(result *TallyResult) bool {
	// all voting power
	votes := result.GetVotesByType(For) + result.GetVotesByType(Abstain) + result.GetVotesByType(Against) + result.GetVotesByType(NoWithVeto)
	vetoThreshold := votes * gvr.gs.getVetoNumerator() / gvr.gs.getVetoDenominator()
	return result.GetVotesByType(NoWithVeto) >= vetoThreshold
}

// TODO: super majority required
func (gvr *Governor) voteSucceeded(result *TallyResult) bool {
	// get all votes count
	votes := result.GetVotesByType(For) + result.GetVotesByType(Abstain) + result.GetVotesByType(Against) + result.GetVotesByType(NoWithVeto)
	succeedThreshold := votes * gvr.gs.getSucceedNumerator() / gvr.gs.getSucceedDenominator()
	return result.GetVotesByType(For) >= succeedThreshold
}

// TODO: undetermined rules
func (gvr *Governor) isBurnDeposit(proposalId string) bool {
	return false
}
