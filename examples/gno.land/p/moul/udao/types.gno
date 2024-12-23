package udao

import (
	"time"
)

// DAO defines the interface for proposal management
type DAO interface {
	// Core proposal operations
	Propose(def PropDefinition) (Proposal, error)
	GetProposal(proposalID uint64) (Proposal, error)
	Execute(proposalID uint64) error

	// List operations
	ActiveProposals() PropList
	ArchivedProposals() PropList
	Len() int
}

// PropDefinition defines the content of a proposal
type PropDefinition interface {
	Title() string
	Body() string
	Created() time.Time
	Finished() time.Time
	CheckConstraints() (valid bool, reasons []string)
}

// Proposal represents a complete proposal including its ID, definition and status
type Proposal interface {
	ID() uint64
	Definition() PropDefinition
	GetState() PropState
	GetYeaPercentage() float64
	GetNayPercentage() float64
	GetAbstainPercentage() float64
	GetPendingVoterPercentage() float64
}

// PropState represents the state of a proposal
type PropState string

const (
	Pending   PropState = "pending"
	Active    PropState = "active"
	Passed    PropState = "passed"
	Failed    PropState = "failed"
	Executed  PropState = "executed"
	Cancelled PropState = "cancelled"
)

// PropList defines the read-only operations available on a proposal list
type PropList interface {
	Len() int
	Get(index int) Proposal
	Slice(startIndex, endIndex int) []Proposal
}
