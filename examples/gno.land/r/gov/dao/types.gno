package dao

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type ProposalID int64

func (pid ProposalID) String() string {
	return ufmt.Sprintf("%v", int64(pid))
} 

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


type ProposalRequestBuilder struct {
	title string
	description string
	callback func(Metadata) error
	metadata Metadata
}

func NewProposalRequestBuilder(title string) *ProposalRequestBuilder {
	return &ProposalRequestBuilder{
		title:title,
	}
}

func (prb *ProposalRequestBuilder) WithDescription(desc string) *ProposalRequestBuilder {
	prb.description = desc
	return prb
}

func (prb *ProposalRequestBuilder) WithMetadata(md Metadata) *ProposalRequestBuilder {
	prb.metadata = md
	return prb
}

func (prb *ProposalRequestBuilder) WithCallback(cb func(Metadata) error) *ProposalRequestBuilder {
	prb.callback = cb
	return prb
}

func (prb *ProposalRequestBuilder) Build() ProposalRequest {
	return ProposalRequest {
		title: prb.title,
		description: prb.description,
		callback: prb.callback,
		metadata: prb.metadata,
	}	
}

type ProposalRequest struct {
	title       string
	description string
	callback    func(Metadata) error
	metadata    Metadata
}

func (p *ProposalRequest) Title() string {
	return p.title
}

func (p *ProposalRequest) Description() string {
	return p.description
}

func (p *ProposalRequest) Callback() func(Metadata)error {
	return p.callback
}

func (p *ProposalRequest) Metadata() Metadata {
	return p.metadata
}

type Proposal struct {
	author Author

	title       string
	description string

	executor    *executor
	allowedDAOs []string
}

func (p *Proposal) Author() Author {
	return p.author
}

func (p *Proposal) Title() string {
	return p.title
}

func (p *Proposal) Description() string {
	return p.description
}

func (p *Proposal) ExecutorString() string {
	if p.executor != nil {
		return p.executor.String()
	}

	return ""
}


func (p *Proposal) AllowedDAOs() []string {
	return p.allowedDAOs
}

type Proposals struct {
	seq int64
	*avl.Tree // *avl.Tree[ProposalID]*Proposal
}

func NewProposals() Proposals {
	return Proposals{Tree:avl.NewTree()}
}

func (ps Proposals) SetProposal(p *Proposal) ProposalID{
	pid := ProposalID(ps.seq)
	ps.Set(pid.String(), p)
	ps.seq++

	return pid
}

func (ps Proposals) GetProposal(pid ProposalID) *Proposal {
	pv, ok := ps.Get(pid.String())
	if !ok {
		return nil
	}

	return pv.(*Proposal)
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
	return func(m Metadata) error {
		// check if the caller is an allowed DAO from the AllowedDAOs list
		if index(std.PreviousRealm().PkgPath()) < 0 {
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
	PreCreateProposal(r ProposalRequest) (Author, error)

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

	// PreExecuteProposal is called when someone is trying to execute a proposal by ID.
	// Is intended to be used to validate who can trigger the proposal execution.
	PreExecuteProposal(pid ProposalID) (bool, error)

	// Render will return a human-readable string in markdown format that
	// will be used to show new data through the dao proxy entrypoint.
	Render(string) string
}

type UpdateRequest struct {
	DAO         DAO
	AllowedDAOs []string
}
