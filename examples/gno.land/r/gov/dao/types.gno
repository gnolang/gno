package dao

import (
	"errors"
	"strings"

	"gno.land/p/nt/bptree/v0"
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

func NewVoteRequest(option VoteOption, proposalID ProposalID) VoteRequest {
	return VoteRequest{
		Option:     option,
		ProposalID: proposalID,
	}
}

func NewVoteRequestWithMetadata(option VoteOption, proposalID ProposalID, metadata interface{}) VoteRequest {
	return VoteRequest{
		Option:     option,
		ProposalID: proposalID,
		Metadata:   metadata,
	}
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
	seq            seqid.ID
	*bptree.BPTree // *bptree.BPTree[ProposalID]*Proposal
}

func NewProposals() *Proposals {
	return &Proposals{BPTree: bptree.NewBPTree32()}
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
	pv := ps.Get(pid.String())
	if pv == nil {
		return nil
	}

	return pv.(*Proposal)
}

type Executor interface {
	Execute(cur realm) error
	String() string
	CreationRealm() string
}

// NewSimpleExecutor constructs an Executor whose creationRealm is captured
// from rlm.PkgPath() at construction time. The IsCurrent() check rejects
// stale or stashed realm values so the captured value is the authentic
// caller realm. creationRealm is display-only (rendered as "Executor
// created in: ..." in proposal listings) — no auth gate downstream.
func NewSimpleExecutor(_ int, rlm realm, callback func(realm) error, description string) *SimpleExecutor {
	if !rlm.IsCurrent() {
		panic("NewSimpleExecutor: rlm is not the caller's live cur (stale capture or sibling frame)")
	}
	if callback == nil {
		panic("executor callback must not be nil")
	}

	return &SimpleExecutor{
		callback:      callback,
		desc:          description,
		creationRealm: rlm.PkgPath(),
	}
}

// SimpleExecutor implements the Executor interface using
// a callback function and a description string.
type SimpleExecutor struct {
	callback      func(realm) error
	desc          string
	creationRealm string
}

// proxyPkgPath is this package: the realm whose frame Execute mints for a
// callback, and therefore the only realm entitled to invoke one.
const proxyPkgPath = "gno.land/r/gov/dao"

// Execute runs the proposal's callback. Invocable only from this proxy.
//
// SECURITY: Execute is a CROSSING method declared here, so invoking it mints
// a gno.land/r/gov/dao frame for the callback — `cur.Previous()` inside the
// callback is this path whoever called. That is a capability: realms gate on
// it (authz.NewContractAuthority("gno.land/r/gov/dao") + DoByPrevious,
// ownable, anything reading Previous().Address()). Ungated, any realm could
// wrap a reachable `func(realm) error` — including an ordinary exported
// entrypoint of the victim — and run it with governance's identity, with no
// proposal id minted, so nothing to render, audit or deny afterwards.
//
// NOT InAllowedDAOs: that is SafeExecutor's bug. It holds the impl path while
// an executor's caller is the proxy, so it rejects the approved route and
// fails open while empty. Exact path plus subpackages, because
// impl.ExecuteProposal is non-crossing (Previous() is the proxy exactly), so
// a bare prefix would reject every real proposal.
//
// Gates the executor object's INVOCATION only. Keeping a privileged closure,
// or an exported entrypoint shaped like `func(realm) error`, out of reach
// stays the consumer's job — see r/gnops/valopers/admin.gno
// and p/moul/authz's NewContractAuthority godoc.
func (e *SimpleExecutor) Execute(cur realm) error {
	// IsCurrent first, as every other gate here does: otherwise Previous() is
	// read from whatever realm value the caller threaded in.
	if !cur.IsCurrent() {
		return errors.New("execution denied: cur is not the caller's live realm")
	}
	if prev := cur.Previous().PkgPath(); prev != proxyPkgPath &&
		!strings.HasPrefix(prev, proxyPkgPath+"/") {
		return errors.New("execution denied: executors are only invocable by " + proxyPkgPath)
	}

	// Check if executor was created using the constructor func
	if e.callback == nil {
		return nil
	}

	return e.callback(cross(cur))
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
	// IsCurrent first, matching every other allowlist gate in this tree
	// (proxy.go's UpdateImpl, treasury, memberstore). Without it this method
	// trusts whatever realm value it is handed, so a caller threading a stale
	// or sibling-frame cur would have its Previous() read from that value
	// rather than from the live frame.
	//
	// NewSafeExecutor has no call sites, so this type is currently dead code.
	// Live proposal execution goes through the Executor interface to
	// SimpleExecutor.Execute, which now gates on the invoking realm being the
	// proxy -- deliberately NOT on InAllowedDAOs, which holds the impl path
	// and therefore rejects the approved route and fails open while empty.
	// That is this method's bug; see SimpleExecutor.Execute.
	if !cur.IsCurrent() {
		return errors.New("execution denied: cur is not the caller's live realm")
	}
	// Verify the caller is an adequate Realm
	if !InAllowedDAOs(cur.Previous().PkgPath()) {
		return errors.New("execution only allowed by validated govDAOs")
	}

	return e.e.Execute(cross(cur))
}

func (e *SafeExecutor) String() string {
	return e.e.String()
}

func (e *SafeExecutor) CreationRealm() string {
	return e.e.CreationRealm()
}

// DAO is the govDAO implementation interface. All mutating/auth-gated
// methods take rlm as their realm-typed parameter in the second position
// (the `_ int, rlm realm` non-crossing form): callers thread the proxy's
// cur as data without forcing a realm transition, so the impl's existing
// unsafe.CurrentRealm()-based auth gates (isValidCall, memberstore.Get)
// continue to see the proxy realm. Render stays unchanged.
type DAO interface {
	// PreCreateProposal is called just before creating a new Proposal
	// It is intended to be used to get the address of the proposal, that
	// may vary depending on the DAO implementation, and to validate that
	// the requester is allowed to do a proposal
	PreCreateProposal(_ int, rlm realm, r ProposalRequest) (address, error)

	// PostCreateProposal is called after creating the Proposal. It is
	// intended to be used as a way to store a new proposal status, that
	// depends on the actuall govDAO implementation
	PostCreateProposal(_ int, rlm realm, r ProposalRequest, pid ProposalID)

	// VoteOnProposal will send a petition to vote for a specific proposal
	// to the actual govDAO implementation
	VoteOnProposal(_ int, rlm realm, r VoteRequest) error

	// PreExecuteProposal is called when someone is trying to execute a proposal by ID.
	// Is intended to be used to validate who can trigger the proposal execution.
	PreExecuteProposal(_ int, rlm realm, pid ProposalID) (bool, error)

	// ExecuteProposal executes the proposal executor and on error changes proposal
	// status to denied with the error message being the denial reason.
	// It returns the executor error when it fails.
	ExecuteProposal(_ int, rlm realm, pid ProposalID, e Executor) error

	// Render will return a human-readable string in markdown format that
	// will be used to show new data through the dao proxy entrypoint.
	// Crossing: the chain query layer auto-injects .cur, and
	// implementations forward cur to internal rlm-aware helpers (mux
	// RenderRlm + downstream cross(rlm) reads).
	Render(cur realm, pkgpath string, path string) string
}

type UpdateRequest struct {
	DAO         DAO
	AllowedDAOs []string
}

// NewUpdateRequest copies allowedDAOs into a fresh slice owned by
// /r/gov/dao. Under the storage=authority model, if we stored the
// caller-passed slice directly, the base ArrayValue would retain
// PkgID = caller_realm: storage rent would attribute to caller, and
// /r/gov/dao could not mutate (e.g. append to) its own copy without
// a DidUpdate panic. The internal copy ensures the UpdateRequest
// and its AllowedDAOs both live entirely in /r/gov/dao's authority.
func NewUpdateRequest(d DAO, allowedDAOs []string) UpdateRequest {
	cp := make([]string, len(allowedDAOs))
	copy(cp, allowedDAOs)
	return UpdateRequest{
		DAO:         d,
		AllowedDAOs: cp,
	}
}
