package dao

// ProposalRequest is a single govdao proposal request
// that contains the necessary information to
// log and generate a valid proposal
type ProposalRequest struct {
	Description string   // the description associated with the proposal
	Executor    Executor // the proposal executor
}

// DAO defines the DAO abstraction
type DAO interface {
	// PropStore is the DAO proposal storage
	PropStore

	// Propose adds a new proposal to the executor-based GOVDAO.
	// Returns the generated proposal ID
	Propose(request ProposalRequest) (uint64, error)

	// VoteOnProposal adds a vote to the given proposal ID
	VoteOnProposal(id uint64, option VoteOption) error

	// ExecuteProposal executes the proposal with the given ID
	ExecuteProposal(id uint64) error
}
