package proxy

import (
	"errors"
	"std"
)
type ProposalID int

// VoteOption is the limited voting option for a DAO proposal
// New govDAOs can create their own VoteOptions if needed in the
// future.
type VoteOption string

const (
	YesVote     VoteOption = "YES"     // Proposal should be accepted
	NoVote      VoteOption = "NO"      // Proposal should be rejected
	AbstainVote VoteOption = "ABSTAIN" // Side is not chosen
)

type VoteRequest struct {
	Option     VoteOption
	ProposalID ProposalID
	Metadata   Metadata
}

type ProposalRequest struct {
	Title string
	Description string
	Callback func(Metadata) error
	Metadata Metadata
}

type Proposal struct {
	Author      Author

	Title       string
	Description string

	Executor    *executor
	AllowedDAOs []string
}

// Author is an interface returned by the current govDAO implementation
// representing the creator of a proposal. It is an interface
// because we don't really know who/what will be allowed to create proposals
// in the future, they can be any entity.
type Author interface {
	ID() string
	String() string
}

type executor struct {
	metadata   Metadata
	callbackFn func(m Metadata) error
}

func newExecutor(fn func(m Metadata) error, metadata Metadata) *executor {
	return &executor{
		metadata:   metadata,
		callbackFn: securityWrapper(fn),
	}
}

func (e *executor) String() string {
	if e.metadata == nil {
		return "No metadata available."
	}

	return e.metadata.String()
}

func (e *executor) Execute() error {
	return e.callbackFn(e.metadata)
}

func securityWrapper(fn func(m Metadata) error) func(m Metadata) error {
	return func (m Metadata) error {
		if index(std.PrevRealm().PkgPath()) < 0 {
			return ErrInvalidExecutorCaller
		}

		return fn(m)
	}
}

var ErrInvalidExecutorCaller error = errors.New("the realm calling the proposal executor was not a govDAO realm")

// Metadata contains any kind of information that will be stored on the executor
// and will be sent to the callback function when the proposal is validated.
// It can contains any kind of data.
type Metadata interface {
	IsMetadata()
	String() string
}

type DAO interface {
	// PreCreateProposal is called just before creating a new Proposal
	// It is intended to be used to get the Author of the proposal, that 
	// may vary depending on the DAO implementation, and to validate that
	// the requester is allowed to do a proposal
	PreCreateProposal(r ProposalRequest)(Author, error)

	// PostCreateProposal is called after creating the Proposal. It is 
    // intended to be used as a way to store a new proposal status, that
	// depends on the actuall govDAO implementation
	PostCreateProposal(r ProposalRequest, pid ProposalID)

	// VoteOnProposal will send a petition to vote for a specific proposal
	// to the actual govDAO implementation
	VoteOnProposal(r VoteRequest) error

	// PreGetProposal is called when someone is trying to get a proposal by ID.
	// Is intended to be used to validate who can query proposals, just in case
	// the actual govDAO implementation wants to limit the access.
	PreGetProposal(pid ProposalID) error

	// PostGetProposal is called after the proposal has been obtained. Intended to be
	// used by govDAO implementations if they need to check Proposal data to know if 
	// the caller is allowed to get that kind of Proposal or not.
	PostGetProposal(pid ProposalID, p *Proposal) error

	// Render will return a human-readable string in markdown format that 
	// will be used to show new data through the dao proxy entrypoint.
	Render(string) string
}

type UpdateRequest struct {
	DAO         DAO
	AllowedDAOs []string
}