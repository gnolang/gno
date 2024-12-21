// Package udao defines minimal interfaces for Decentralized Autonomous Organizations (DAOs).
// It intentionally does not expose members and votes, as these details are implementation-specific.
// Instead, it focuses on providing an external view of proposals and their statuses,
// which is what non-members and non-voters typically care about.
//
// The package is designed to allow for composable DAO patterns, enabling flexible
// and modular implementations of various DAO structures and behaviors.
package udao

// DAO defines a minimal interface for a Decentralized Autonomous Organization
// from an external point of view, hiding the internal details of members and voting.
type DAO interface {
	// Propose submits a new proposal to the DAO
	Propose(proposal Proposal) (uint64, error)

	// GetProposalStatus retrieves the current status and metrics of a specific proposal
	GetProposalStatus(proposalID uint64) (ProposalStatus, error)

	// Execute attempts to execute a proposal if it has passed
	Execute(proposalID uint64) error

	// GetProposal retrieves the details of a specific proposal
	GetProposal(proposalID uint64) (Proposal, error)

	// XXX: find a smart way to list proposals
	// // ListProposalss retrieves a list of Proposals with pagination and optional filters
	// ListDAOs(offset, limit int, filters ...ProposalFilter) ([]DAO, error)
}

// Proposal defines the interface for a DAO proposal
type Proposal interface {
	Title() string
	Body() string
	Constraints() []Constraint
}

// ProposalStatus represents the current status and metrics of a proposal
type ProposalStatus struct {
	// status
	State ProposalState

	// metrics
	YeaPercentage      float64
	NayPercentage      float64
	AbstainPercentage float64
	PendingVoterPercentage float64
	// XXX: other metrics?
}

// ProposalState represents the current state of a proposal
type ProposalState int

const (
	Pending ProposalState = iota
	Active
	Passed
	Rejected
	Executed
	Expired // XXX: better name for when it's expired but not because of time?
)

// Constraint defines an interface for proposal constraints
type Constraint interface {
	Validate() (bool, string)
	Description() string
}

// ValidateConstraints checks if all constraints of a proposal are met
func ValidateConstraints(p Proposal) (bool, []string) {
	var unmetReasons []string
	for _, constraint := range p.Constraints() {
		valid, reason := constraint.Validate()
		if !valid {
			unmetReasons = append(unmetReasons, reason)
		}
	}
	return len(unmetReasons) == 0, unmetReasons
}
