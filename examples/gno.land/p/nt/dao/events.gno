package dao

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// EmitProposalAdded emits an event signaling that
// a given proposal was added
func EmitProposalAdded(id uint64, proposer std.Address) {
	std.Emit(
		ProposalAddedEvent,
		ProposalEventIDKey, ufmt.Sprintf("%d", id),
		ProposalEventAuthorKey, proposer.String(),
	)
}

// EmitProposalAccepted emits an event signaling that
// a given proposal was accepted
func EmitProposalAccepted(id uint64) {
	std.Emit(
		ProposalAcceptedEvent,
		ProposalEventIDKey, ufmt.Sprintf("%d", id),
	)
}

// EmitProposalNotAccepted emits an event signaling that
// a given proposal was not accepted
func EmitProposalNotAccepted(id uint64) {
	std.Emit(
		ProposalNotAcceptedEvent,
		ProposalEventIDKey, ufmt.Sprintf("%d", id),
	)
}

// EmitProposalExecuted emits an event signaling that
// a given proposal was executed, with the given status
func EmitProposalExecuted(id uint64, status ProposalStatus) {
	std.Emit(
		ProposalExecutedEvent,
		ProposalEventIDKey, ufmt.Sprintf("%d", id),
		ProposalEventExecutionKey, status.String(),
	)
}

// EmitVoteAdded emits an event signaling that
// a vote was cast for a given proposal
func EmitVoteAdded(id uint64, voter std.Address, option VoteOption) {
	std.Emit(
		VoteAddedEvent,
		VoteAddedIDKey, ufmt.Sprintf("%d", id),
		VoteAddedAuthorKey, voter.String(),
		VoteAddedOptionKey, option.String(),
	)
}
