package dao

import (
	"chain/runtime"
	"errors"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/seqid/v0"
)

type ProposalID int64

func (pid ProposalID) String() string {
	return seqid.ID(pid).String()
}

// VoteOption is the limited voting option for a DAO proposal
// New govDAOs can create their own VoteOptions if needed in the
// future.
type VoteOption string

const (
	AbstainVote VoteOption = "ABSTAIN" // Side is not chosen
	YesVote     VoteOption = "YES"     // Proposal should be accepted
	NoVote      VoteOption = "NO"      // Proposal should be rejected
)

type VoteRequest struct {
	Option     VoteOption
	ProposalID ProposalID
	Metadata   interface{}
}

func NewProposalRequest(title string, description string, executor Executor) ProposalRequest {
	return ProposalRequest{
		title:       title,
		description: description,
		executor:    executor,
	}
}

func NewProposalRequestWithFilter(title string, description string, executor Executor, filter Filter) ProposalRequest {
	return ProposalRequest{
		title:       title,
		description: description,
		executor:    executor,
		filter:      filter,
	}
}

type Filter interface{}

type ProposalRequest struct {
	title       string
	description string
	executor    Executor
	filter      Filter
}

func (p *ProposalRequest) Title() string {
	return p.title
}

func (p *ProposalRequest) Description() string {
	return p.description
}

func (p *ProposalRequest) Filter() Filter {
	return p.filter
}

type Proposal struct {
	author address

	title       string
	description string

	executor    Executor
	allowedDAOs []string
}

func (p *Proposal) Author() address {
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

func (p *Proposal) ExecutorCreationRealm() string {
	if p.executor != nil {
		return p.executor.CreationRealm()
	}

	return ""
}

func (p *Proposal) AllowedDAOs() []string {
	return append([]string(nil), p.allowedDAOs...)
}

type Proposals struct {
	seq       seqid.ID
	*avl.Tree // *avl.Tree[ProposalID]*Proposal
}

func NewProposals() *Proposals {
	return &Proposals{Tree: avl.NewTree()}
}

func (ps *Proposals) SetProposal(p *Proposal) ProposalID {
	pid := ProposalID(int64(ps.seq))
	updated := ps.Set(pid.String(), p)
	if updated {
		panic("fatal error: Override proposals is not allowed")
	}
	ps.seq = ps.seq.Next()
	return pid
}

func (ps *Proposals) GetProposal(pid ProposalID) *Proposal {
	pv, ok := ps.Get(pid.String())
	if !ok {
		return nil
	}

	return pv.(*Proposal)
}

type Executor interface {
	Execute(cur realm) error
	String() string
	CreationRealm() string
}

func NewSimpleExecutor(callback func(realm) error, description string) *SimpleExecutor {
	if callback == nil {
		panic("executor callback must not be nil")
	}

	return &SimpleExecutor{
		callback:      callback,
		desc:          description,
		creationRealm: runtime.CurrentRealm().PkgPath(),
	}
}

// SimpleExecutor implements the Executor interface using
// a callback function and a description string.
type SimpleExecutor struct {
	callback      func(realm) error
	desc          string
	creationRealm string
}

func (e *SimpleExecutor) Execute(cur realm) error {
	// Check if executor was created using the constructor func
	if e.callback == nil {
		return nil
	}

	return e.callback(cross)
}

func (e *SimpleExecutor) String() string {
	return e.desc
}

func (e *SimpleExecutor) CreationRealm() string {
	return e.creationRealm
}

func NewSafeExecutor(e Executor) *SafeExecutor {
	return &SafeExecutor{
		e: e,
	}
}

// SafeExecutor wraps an Executor to only allow its execution
// by allowed govDAOs.
type SafeExecutor struct {
	e Executor
}

func (e *SafeExecutor) Execute(cur realm) error {
	// Verify the caller is an adequate Realm
	if !InAllowedDAOs(runtime.PreviousRealm().PkgPath()) {
		return errors.New("execution only allowed by validated govDAOs")
	}

	return e.e.Execute(cross)
}

func (e *SafeExecutor) String() string {
	return e.e.String()
}

func (e *SafeExecutor) CreationRealm() string {
	return e.e.CreationRealm()
}

type DAO interface {
	// PreCreateProposal is called just before creating a new Proposal
	// It is intended to be used to get the address of the proposal, that
	// may vary depending on the DAO implementation, and to validate that
	// the requester is allowed to do a proposal
	PreCreateProposal(r ProposalRequest) (address, error)

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
	Render(pkgpath string, path string) string
}

type UpdateRequest struct {
	DAO         DAO
	AllowedDAOs []string
}
