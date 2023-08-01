package governor

import std "std"

type GovernorSettings struct {
	votingDelay       int64
	votingPeriod      int64
	minDeposit        std.Coin // minimal deposit to propose
	proposalThreshold int64    // minimal voting power required to propose
	quorumNumerator   uint64
	succeedNumerator  uint64
	vetoNumerator     uint64
}

func NewGovernorSettings(votingDelay int64, votingPeriod int64, minDeposit std.Coin, proposalThreshold int64, qn, sn, vn uint64) *GovernorSettings {
	return &GovernorSettings{
		votingDelay:       votingDelay,
		votingPeriod:      votingPeriod,
		proposalThreshold: proposalThreshold,
		minDeposit:        minDeposit,
		quorumNumerator:   qn,
		succeedNumerator:  sn,
		vetoNumerator:     vn,
	}
}

type VotingDelaySetEvent struct {
	oldVotingDelay int64
	newVotingDelay int64
}

type VotingPeriodSetEvent struct {
	oldVotingPeriod int64
	newVotingPeriod int64
}

type ProposalThresholdSetEvent struct {
	oldProposalThreshold int64
	newProposalThreshold int64
}

func (gs *GovernorSettings) getVotingDelay() int64 {
	return gs.votingDelay
}

func (gs *GovernorSettings) getVotingPeriod() int64 {
	return gs.votingPeriod
}

func (gs *GovernorSettings) getProposalThreshold() int64 {
	return gs.proposalThreshold
}

// Update the voting delay. This operation can only be performed through a governance proposal.
func (gs *GovernorSettings) setVotingDelay(newVotingDelay int64) {
	caller := std.GetOrigCaller()
	if caller != admin {
		panic("only governance")
	}
	gs.votingDelay = newVotingDelay
}

// Update the voting period. This operation can only be performed through a governance proposal.
func (gs *GovernorSettings) setVotingPeriod(newVotingPeriod int64) {
	caller := std.GetOrigCaller()
	if caller != admin {
		panic("only governance")
	}
	// voting period must be at least one block long
	if newVotingPeriod <= 0 {
		panic("GovernorSettings: voting period too low")
	}
	// TODO: emit VotingPeriodSetEvent
	gs.votingPeriod = newVotingPeriod
}

// Update the proposal threshold. This operation can only be performed through a governance proposal.
func (gs *GovernorSettings) setProposalThreshold(newProposalThreshold int64) {
	caller := std.GetOrigCaller()
	if caller != admin {
		panic("only governance")
	}
	gs.proposalThreshold = newProposalThreshold
	// TODO: emit ProposalThresholdSetEvent
}

// -------------------------------------------
type QuorumNumeratorUpdatedEvent struct {
	oldQuorumNumerator uint64
	newQuorumNumerator uint64
}

// Returns the quorum for a block number, in terms of number of votes: `supply * numerator / denominator`.
func (gs *GovernorSettings) quorum(totalSupply uint64) uint64 {
	return (totalSupply * gs.getQuorumNumerator()) / gs.getQuorumDenominator()
}

// Returns the current quorum numerator
func (gs *GovernorSettings) getQuorumNumerator() uint64 {
	return gs.quorumNumerator
}

// Returns the quorum denominator. Defaults to 100.
func (gs *GovernorSettings) getQuorumDenominator() uint64 {
	return 100
}

// Requirements:
// - Must be called through a governance proposal.
// - New numerator must be smaller or equal to the denominator.
func (gs *GovernorSettings) updateQuorumNumerator(newQuorumNumerator uint64) {
	if newQuorumNumerator > gs.getQuorumDenominator() {
		panic("GovernorVotesQuorumFraction: quorumNumerator over quorumDenominator")
	}
	// oldQuorumNumerator := gvqf.quorumNumerator
	gs.quorumNumerator = newQuorumNumerator
	// emit QuorumNumeratorUpdated(oldQuorumNumerator, newQuorumNumerator);
}

// -----------------------------------------------

type SucceedNumeratorUpdatedEvent struct {
	oldSucceedNumerator uint64
	newSucceedNumerator uint64
}

// Returns the current quorum numerator
func (gs *GovernorSettings) getSucceedNumerator() uint64 {
	return gs.succeedNumerator
}

// Returns the quorum denominator. Defaults to 100.
func (gs *GovernorSettings) getSucceedDenominator() uint64 {
	return 3
}

func (gs *GovernorSettings) updateSucceedNumerator(newSucceedNumerator uint64) {
	if newSucceedNumerator > gs.getSucceedDenominator() {
		panic("GovernorVotesQuorumFraction: quorumNumerator over quorumDenominator")
	}
	// oldQuorumNumerator := gvqf.quorumNumerator
	gs.succeedNumerator = newSucceedNumerator
	// emit QuorumNumeratorUpdated(oldQuorumNumerator, newQuorumNumerator);
}

// ----------------------------------------------------
type VetoNumeratorUpdatedEvent struct {
	oldVetoNumerator uint64
	newVetoNumerator uint64
}

// Returns the current quorum numerator
func (gs *GovernorSettings) getVetoNumerator() uint64 {
	return gs.vetoNumerator
}

// Returns the quorum denominator. Defaults to 100.
func (gs *GovernorSettings) getVetoDenominator() uint64 {
	return 3
}

func (gs *GovernorSettings) updateVetoNumerator(newVetoNumerator uint64) {
	if newVetoNumerator > gs.getVetoDenominator() {
		panic("GovernorVotesQuorumFraction: quorumNumerator over quorumDenominator")
	}
	// oldQuorumNumerator := gvqf.quorumNumerator
	gs.vetoNumerator = newVetoNumerator
	// emit QuorumNumeratorUpdated(oldQuorumNumerator, newQuorumNumerator);
}
