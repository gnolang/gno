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
	// GetProposals returns the given paginated proposals
	GetProposals(offset, count uint64) []Proposal

	// GetProposalByID returns the proposal associated with
	// the given ID, if any
	GetProposalByID(id uint64) (Proposal, error)

	// GetProposalsByAddress returns the proposals associated
	// with the given proposer address
	GetProposalsByAddress(address std.Address) []Proposal
}

// Proposal is the single proposal abstraction
type Proposal interface {
	// GetAuthor returns the author of the proposal
	GetAuthor() std.Address

	// GetDescription returns the description of the proposal
	GetDescription() string

	// GetStatus returns the status of the proposal
	GetStatus() ProposalStatus

	// GetVotes returns the votes of the proposal
	GetVotes() []Vote

	// GetVoteByMember returns the proposal vote by the member, if any
	GetVoteByMember(address std.Address) (Vote, error)
}
