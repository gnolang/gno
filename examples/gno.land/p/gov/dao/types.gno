package dao

import (
	"std"

	"gno.land/p/gov/proposal"
)

type VoteOption string

const (
	YesVote VoteOption = "YES"
	NoVote  VoteOption = "NO"
	Abstain  VoteOption = "ABSTAIN"
)

func (v VoteOption) String() string {
	return string(v)
}

// Vote encompasses a single GOVDAO vote
type Vote struct {
	Address std.Address // the address of the voter
	Option  VoteOption  // the voting option
}

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
