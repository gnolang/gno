// Package udao provides composable interfaces for DAO (Decentralized Autonomous Organization) implementations.
//
// The package focuses on proposal management with minimal, composable interfaces that allow:
// 1. Composable DAOs where proposals can draw voting power from multiple DAO sources
// 2. Interchain compatibility by focusing on top-level metrics rather than internal details
// 3. Rich feature implementations while maintaining a minimal common interface
//
// Core Interfaces:
//
// DAO: The main interface for proposal management including creation, retrieval, and execution
//
// PropDefinition: Defines the content and constraints of a proposal
//
// Proposal: Represents a complete proposal with ID, definition, state, and voting percentages
//
// PropList: Read-only interface for accessing collections of proposals
//
// Example Usage:
//
// A basic implementation is provided in the basicdao subpackage:
//
//	dao := basicdao.NewBasicDAO()
//	proposal := basicdao.NewSimpleProposal("My Proposal", "Proposal description")
//	prop, err := dao.Propose(proposal)
//	if err != nil {
//		// handle error
//	}
//	
//	// Check voting status
//	println("Yea:", prop.GetYeaPercentage())
//	println("State:", prop.GetState())
//
// The interfaces are designed to be composable, allowing multiple DAO implementations
// to coexist and interact while maintaining their own internal voting mechanisms.
package udao

import "time"

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
