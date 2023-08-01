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
	return gvr.gs.quorum(gvr.govToken.TotalSupply()) <= totalVotingPower
}

// veto reached, 33.4
func (gvr *Governor) vetoReached(result *TallyResult) bool {
	// all voting power
	votes := result.GetVotes(For) + result.GetVotes(Abstain) + result.GetVotes(Against) + result.GetVotes(NoWithVeto)
	vetoThreshold := votes * gvr.gs.getVetoNumerator() / gvr.gs.getVetoDenominator()
	return result.GetVotes(NoWithVeto) >= vetoThreshold
}

// TODO: super majority required
func (gvr *Governor) voteSucceeded(result *TallyResult) bool {
	// get all votes count
	votes := result.GetVotes(For) + result.GetVotes(Abstain) + result.GetVotes(Against) + result.GetVotes(NoWithVeto)
	succeedThreshold := votes * gvr.gs.getSucceedNumerator() / gvr.gs.getSucceedDenominator()
	return result.GetVotes(For) >= succeedThreshold
}

// TODO: undetermined rules
func (gvr *Governor) isBurnDeposit(proposalId string) bool {
	return false
}
