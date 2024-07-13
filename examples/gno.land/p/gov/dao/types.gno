package dao

import "gno.land/p/gov/proposal"

type VoteOption string

const (
	YesVote VoteOption = "YES"
	NoVote  VoteOption = "NO"
)

func (v VoteOption) String() string {
	return string(v)
}

// DAO defines the DAO abstraction
type DAO interface {
	// Propose adds a new proposal to the executor-based GOVDAO
	Propose(comment string, executor proposal.Executor) error

	// VoteOnProposal adds a vote to the given proposal ID
	VoteOnProposal(id uint64, option VoteOption) error

	// ExecuteProposal executes the proposal with the given ID
	ExecuteProposal(id uint64) error

	// GetMembersStore returns the member store associated with the DAO
	GetMembersStore() MemberStore

	// SetMembersStore sets the member store for the DAO
	SetMembersStore(store MemberStore)

	// GetPropStore returns the proposal store associated with the DAO
	GetPropStore() PropStore

	// SetPropStore sets the proposal store for the DAO
	SetPropStore(store PropStore)
}
