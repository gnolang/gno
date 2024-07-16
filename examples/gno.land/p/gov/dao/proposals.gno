package dao

import "std"

type ProposalStatus string

// ACTIVE -> ACCEPTED -> EXECUTION(SUCCEEDED/FAILED)
// ACTIVE -> NOT ACCEPTED
var (
	Active              ProposalStatus = "active"               // proposal is still active
	Accepted            ProposalStatus = "accepted"             // proposal gathered quorum
	NotAccepted         ProposalStatus = "not accepted"         // proposal failed to gather quorum
	ExecutionSuccessful ProposalStatus = "execution successful" // proposal is executed successfully
	ExecutionFailed     ProposalStatus = "execution failed"     // proposal is failed during execution
)

func (s ProposalStatus) String() string {
	return string(s)
}

// PropStore defines the proposal storage abstraction
type PropStore interface {
	// Proposals returns the given paginated proposals
	Proposals(offset, count uint64) []Proposal

	// ProposalByID returns the proposal associated with
	// the given ID, if any
	ProposalByID(id uint64) (Proposal, error)

	// Size returns the number of proposals in
	// the proposal store
	Size() int
}

// Proposal is the single proposal abstraction
type Proposal interface {
	// Author returns the author of the proposal
	Author() std.Address

	// Description returns the description of the proposal
	Description() string

	// Status returns the status of the proposal
	Status() ProposalStatus

	// Votes returns the votes of the proposal
	Votes(offset, count uint64) []Vote

	// VotingStats returns the voting stats of the proposal
	VotingStats() VotingStats

	// VoteByMember returns the proposal vote by the member, if any
	VoteByMember(address std.Address) (Vote, error)
}
