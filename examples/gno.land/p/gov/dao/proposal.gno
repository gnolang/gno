package dao

import "std"

type Status string

// ACTIVE -> ACCEPTED -> EXECUTED
// ACTIVE -> NOT ACCEPTED
var (
	Active      Status = "active"       // proposal is still active
	Accepted    Status = "accepted"     // proposal gathered quorum
	NotAccepted Status = "not accepted" // proposal failed to gather quorum
	Executed    Status = "executed"     // proposal is executed
)

func (s Status) String() string {
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
	GetStatus() Status

	// GetVotes returns the votes of the proposal
	GetVotes() []Vote
}
