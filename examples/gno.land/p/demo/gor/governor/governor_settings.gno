package governor

import "std"
import "gno.land/p/demo/maths"

type GovernorSettings struct {
	votingDelay  int64
	votingPeriod int64
	//minDeposit        std.Coin // minimal deposit to propose
	proposalThreshold maths.Rat // minimal voting power required to propose
	quorum            maths.Rat
	succeedThreshold  maths.Rat
	vetoThreshold     maths.Rat
}

func NewGovernorSettings(votingDelay int64, votingPeriod int64, proposalThreshold maths.Rat, qum, st, vt maths.Rat) *GovernorSettings {
	return &GovernorSettings{
		votingDelay:       votingDelay,
		votingPeriod:      votingPeriod,
		proposalThreshold: proposalThreshold,
		//minDeposit:        minDeposit,
		quorum:           qum,
		succeedThreshold: st,
		vetoThreshold:    vt,
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
	proposalThreshold maths.Rat
}

func (gs *GovernorSettings) getVotingDelay() int64 {
	return gs.votingDelay
}

func (gs *GovernorSettings) getVotingPeriod() int64 {
	return gs.votingPeriod
}

func (gs *GovernorSettings) getProposalThreshold() maths.Rat {
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
func (gs *GovernorSettings) setProposalThreshold(newProposalThreshold maths.Rat) {
	caller := std.GetOrigCaller()
	if caller != admin {
		panic("only governance")
	}
	gs.proposalThreshold = newProposalThreshold
	// TODO: emit ProposalThresholdSetEvent
}

// -------------------------------------------
type QuorumUpdatedEvent struct {
	quorum maths.Rat
}

// Returns the quorum for a block number, in terms of number of votes: `supply * numerator / denominator`.
func (gs *GovernorSettings) getQuorum() maths.Rat {
	return gs.quorum
}

// -----------------------------------------------

type SucceedThresholdUpdatedEvent struct {
	succeedThreshold maths.Rat
}

// Returns the current quorum numerator
func (gs *GovernorSettings) getSucceedThreshold() maths.Rat {
	return gs.succeedThreshold
}

// ----------------------------------------------------
type VetoThresholdUpdatedEvent struct {
	vetoThreshold maths.Rat
}

// Returns the current quorum numerator
func (gs *GovernorSettings) getVetoThreshold() maths.Rat {
	return gs.vetoThreshold
}
