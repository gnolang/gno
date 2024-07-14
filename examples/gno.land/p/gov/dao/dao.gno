package dao

import "gno.land/p/gov/proposal"

// DAO defines the DAO abstraction
type DAO interface {
	// Propose adds a new proposal to the executor-based GOVDAO.
	// Returns the generated proposal ID
	Propose(comment string, executor proposal.Executor) (uint64, error)

	// VoteOnProposal adds a vote to the given proposal ID
	VoteOnProposal(id uint64, option VoteOption) error

	// ExecuteProposal executes the proposal with the given ID
	ExecuteProposal(id uint64) error

	// GetMembStore returns the member store associated with the DAO
	GetMembStore() MemberStore

	// GetPropStore returns the proposal store associated with the DAO
	GetPropStore() PropStore
}
