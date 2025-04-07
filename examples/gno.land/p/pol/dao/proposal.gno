package dao

import "time"

// A proposal status is an enum
type ProposalChoice uint

// A general implementation for an executor
type Executor interface {
	// Does the action
	Execute() error
}

// A general implementation for a proposal
type Proposal interface {
	// Get the author of the proposal
	GetAuthor() Member

	// Check if the proposal is accepted
	IsAccepted() bool

	// Execute if the proposal has been accepted
	ExecuteIfAccepted() error

	// When does it expires
	ExpiresIn() time.Time

	// Make someone vote
	Vote(verdict ProposalChoice) error

	// Check out how many people voted
	HowManyVoters() uint

	// Check if its ended
	IsEnded() bool
}

type ProposalStore interface {
	// List all proposals
	GetProposals() []Proposal

	// Propose something to the group
	Propose(prop Proposal)

	// Get the choices names & count
	GetChoicesNames() map[ProposalChoice]string
}
