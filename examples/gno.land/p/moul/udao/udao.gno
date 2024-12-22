// Package udao defines minimal interfaces for Decentralized Autonomous Organizations (DAOs).
// It intentionally does not expose members and votes, as these details are implementation-specific.
// Instead, it focuses on providing an external view of proposals and their statuses,
// which is what non-members and non-voters typically care about.
//
// The package is designed to allow for composable DAO patterns, enabling flexible
// and modular implementations of various DAO structures and behaviors.
package udao

import (
	"time"

	"gno.land/p/demo/avl/rolist"
)

//-----------------------------------------------------------------------------
// DAO Interface
//-----------------------------------------------------------------------------

// DAO defines a minimal interface for a Decentralized Autonomous Organization
type DAO interface {
	// Propose submits a new proposal to the DAO and returns the created Proposal
	Propose(definition PropDefinition) (Proposal, error)

	// GetProposal retrieves a specific proposal
	GetProposal(proposalID uint64) (Proposal, error)

	// Execute attempts to execute a proposal if it has passed
	Execute(proposalID uint64) error

	// ActiveProposals returns a read-only list of active proposals
	ActiveProposals() PropList

	// ArchivedProposals returns a read-only list of archived proposals
	ArchivedProposals() PropList
}

// Proposal represents a complete proposal including its ID, definition and status
type Proposal interface {
	ID() uint64
	Definition() PropDefinition
	PropStatus
}

//-----------------------------------------------------------------------------
// Proposal Types & Helpers
//-----------------------------------------------------------------------------

// PropDefinition defines the interface for a DAO proposal
type PropDefinition interface {
	Title() string
	Body() string
	Created() time.Time
	Constraints() []Constraint
}

// PropStatus represents the current status and metrics of a proposal
type PropStatus struct {
	// status
	State PropState

	// metrics
	YeaPercentage          float64
	NayPercentage          float64
	AbstainPercentage      float64
	PendingVoterPercentage float64
	// XXX: other metrics?
}

// PropState represents the current state of a proposal
type PropState string

const (
	Pending  PropState = "pending"
	Active   PropState = "active"
	Passed   PropState = "passed"
	Rejected PropState = "rejected"
	Executed PropState = "executed"
	Expired  PropState = "expired"
)

// Constraint defines an interface for proposal constraints
type Constraint interface {
	Validate() (valid bool, description string)
}

// ValidateConstraints checks if all constraints of a proposal are met
func ValidateConstraints(p Proposal) (bool, []string) {
	reasons := make([]string, 0)
	for _, c := range p.Definition().Constraints() {
		if valid, reason := c.Validate(); !valid {
			reasons = append(reasons, reason)
		}
	}
	return len(reasons) == 0, reasons
}

//-----------------------------------------------------------------------------
// PropList Interface & Implementation
//-----------------------------------------------------------------------------

// PropList defines the read-only operations available on a proposal list
type PropList interface {
	Len() int
	Get(index int) Proposal
	Slice(startIndex, endIndex int) []Proposal
}

// PropListWrapper wraps an IReadOnlyList to implement PropList
type PropListWrapper struct {
	list rolist.IReadOnlyList
}

// WrapAsPropList converts an IReadOnlyList to a PropList
func WrapAsPropList(list rolist.IReadOnlyList) PropList {
	return &PropListWrapper{list: list}
}

// Len returns the number of proposals in the list
func (pl *PropListWrapper) Len() int {
	return pl.list.Len()
}

// Get returns the proposal and status at the specified index
func (pl *PropListWrapper) Get(index int) Proposal {
	value := pl.list.Get(index)
	if value == nil {
		return nil
	}
	return value.(Proposal)
}

// Slice returns a slice of proposals from startIndex to endIndex
func (pl *PropListWrapper) Slice(startIndex, endIndex int) []Proposal {
	values := pl.list.Slice(startIndex, endIndex)
	if values == nil {
		return nil
	}

	result := make([]Proposal, len(values))
	for i, v := range values {
		result[i] = v.(Proposal)
	}
	return result
}
