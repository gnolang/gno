// Package udao defines minimal interfaces for Decentralized Autonomous Organizations (DAOs).
// It intentionally does not expose members and votes, as these details are implementation-specific.
// Instead, it focuses on providing an external view of proposals and their statuses,
// which is what non-members and non-voters typically care about.
//
// The package is designed to allow for composable DAO patterns, enabling flexible
// and modular implementations of various DAO structures and behaviors.
package udao

import "time"

// DAO defines a minimal interface for a Decentralized Autonomous Organization
type DAO interface {
	// Propose submits a new proposal to the DAO
	Propose(proposal PropDefinition) (uint64, error)

	// GetProposal retrieves both the definition and status of a specific proposal
	GetProposal(proposalID uint64) (PropDefinition, PropStatus, error)

	// Execute attempts to execute a proposal if it has passed
	Execute(proposalID uint64) error

	// XXX: find a smart way to list proposals
}

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
	for _, c := range p.Constraints() {
		if valid, reason := c.Validate(); !valid {
			reasons = append(reasons, reason)
		}
	}
	return len(reasons) == 0, reasons
}
