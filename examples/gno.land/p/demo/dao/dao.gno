package dao

const (
	ProposalAddedEvent       = "ProposalAdded"       // emitted when a new proposal has been added
	ProposalAcceptedEvent    = "ProposalAccepted"    // emitted when a proposal has been accepted
	ProposalNotAcceptedEvent = "ProposalNotAccepted" // emitted when a proposal has not been accepted
	ProposalExecutedEvent    = "ProposalExecuted"    // emitted when a proposal has been executed

	ProposalEventIDKey        = "proposal-id"
	ProposalEventExecutionKey = "exec-status"
)

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

	// ExecuteProposal executes the proposal with the given ID
	ExecuteProposal(id uint64) error
}
