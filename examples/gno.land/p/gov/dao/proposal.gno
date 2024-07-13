package dao

import "std"

type Status string

var (
	Accepted    Status = "accepted"
	Active      Status = "active"
	NotAccepted Status = "not accepted"
	Expired     Status = "expired"
	Succeeded   Status = "succeeded"
)

func (s Status) String() string {
	return string(s)
}

// PropStore defines the proposal storage abstraction
type PropStore interface {
	// AddProposal adds a new proposal to the store
	AddProposal(proposal Proposal) error

	// RemoveProposal removes a proposal from the store, if any
	RemoveProposal(id uint64) error

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

	// GetVoters returns the voters of the proposal
	GetVoters() []std.Address
}
