package dao

// DAO defines the DAO abstraction
type DAO interface {
	// Propose adds a new proposal to the executor-based GOVDAO.
	// Returns the generated proposal ID
	Propose(comment string, executor Executor) (uint64, error)

	// ExecuteProposal executes the proposal with the given ID
	ExecuteProposal(id uint64) error
}
