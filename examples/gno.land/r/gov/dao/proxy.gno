package dao

import (
	"chain"
	"errors"
	"strconv"

	"gno.land/p/nt/ufmt/v0"
)

// dao is the actual govDAO implementation, having all the needed business logic
var dao DAO

// allowedDAOs contains realms that can be used to update the actual govDAO implementation,
// and validate Proposals.
// This is like that to be able to rollback using a previous govDAO implementation in case
// the latest implementation has a breaking bug. After a test period, a proposal can be
// executed to remove all previous govDAOs implementations and leave the last one.
var allowedDAOs []string

// proposals contains all the proposals in history.
var proposals *Proposals = NewProposals()

// Named, exported sentinel errors for VoteOnProposal failure modes (see
// #3104 / #5998). These are declared here, at the dao package's stable
// entry point, rather than in a specific implementation (e.g.
// gno.land/r/gov/dao/v3/impl), because UpdateImpl can swap the underlying
// implementation behind VoteOnProposal at any time: a sentinel declared in
// v3/impl would stop matching via errors.Is as soon as a v4 implementation
// (with its own, distinct sentinel values) took over, even if the error
// message stayed identical. Callers should check errors.Is(err,
// dao.ErrAlreadyVoted) etc. against these values, not against an
// implementation package's own copy.
var (
	ErrMemberNotFound    = errors.New("member not found")
	ErrProposalNotFound  = errors.New("proposal not found")
	ErrNotAllowedToVote  = errors.New("member on specified tier is not allowed to vote on this proposal")
	ErrAlreadyVoted      = errors.New("already voted on proposal")
	ErrInvalidVoteOption = errors.New("voting can only be YES, NO, or ABSTAIN")

	// ErrProposalClosed is the sentinel to check with errors.Is(err,
	// dao.ErrProposalClosed); the concrete error returned is a
	// *ProposalClosedError, which also carries whether the proposal was
	// accepted or denied.
	ErrProposalClosed = errors.New("proposal closed")

	// ErrDirectCallRequired is returned when VoteOnProposal is invoked by
	// a realm rather than directly by a user (or a maketx run script).
	ErrDirectCallRequired = errors.New("proposal voting must be done directly by a user")
)

// ProposalClosedError indicates a vote was attempted on a proposal that
// has already been accepted or denied. It implements Is(error) bool (the
// extension point documented by the errors package) so
// errors.Is(err, ErrProposalClosed) matches regardless of which outcome
// the proposal reached.
type ProposalClosedError struct {
	Accepted bool
}

func (e *ProposalClosedError) Error() string {
	return ufmt.Sprintf("proposal closed. Accepted: %v", e.Accepted)
}

func (e *ProposalClosedError) Is(target error) bool {
	return target == ErrProposalClosed
}

// NewProposalClosedError returns a *ProposalClosedError. Implementations
// (e.g. gno.land/r/gov/dao/v3/impl) must construct it through this
// constructor rather than a struct literal: gno only allows a realm to
// allocate its own types, so `&ProposalClosedError{...}` written in a
// different realm's source panics with "cannot allocate
// gno.land/r/gov/dao.ProposalClosedError in realm ...". Calling this
// function keeps the allocation in dao's own code.
func NewProposalClosedError(accepted bool) *ProposalClosedError {
	return &ProposalClosedError{Accepted: accepted}
}

// Render calls directly to Render's DAO implementation.
// This allows to have this realm as the main entry point for everything.
func Render(cur realm, p string) string {
	if dao == nil {
		return "DAO not initialized"
	}
	return dao.Render(cross(cur), cur.PkgPath(), p)
}

// MustCreateProposal is an utility method that does the same as CreateProposal,
// but instead of erroing if something happens, it panics.
func MustCreateProposal(cur realm, r ProposalRequest) ProposalID {
	pid, err := CreateProposal(cur, r)
	if err != nil {
		panic(err.Error())
	}

	return pid
}

// ExecuteProposal will try to execute the proposal with the provided ProposalID.
// If the proposal was denied, it will return false. If the proposal is correctly
// executed, it will return true. If something happens this function will panic.
func ExecuteProposal(cur realm, pid ProposalID) bool {
	return executeProposal(cur, pid, false)
}

// ExecuteOrRejectProposal executes the proposal with the provided ProposalID or rejects
// it when there is an execution error.
// If the proposal was denied, it will return false. If the proposal is correctly
// executed, it will return true, unless execution fails with an error, in which case
// proposal is rejected with the error as the reason.
// This function allows to finish proposals by rejecting them when there is a state
// change or an error in the proposal parameters that makes execution fail, potentially
// leaving the proposal active forever because it can't be successfully executed.
func ExecuteOrRejectProposal(cur realm, pid ProposalID) bool {
	return executeProposal(cur, pid, true)
}

// CreateProposal will try to create a new proposal, that will be validated by the actual
// govDAO implementation. If the proposal cannot be created, an error will be returned.
func CreateProposal(cur realm, r ProposalRequest) (ProposalID, error) {
	if dao == nil {
		return -1, errors.New("DAO not initialized")
	}
	author, err := dao.PreCreateProposal(0, cur, r)
	if err != nil {
		return -1, err
	}

	p := &Proposal{
		author:      author,
		title:       r.title,
		description: r.description,
		executor:    r.executor,
		allowedDAOs: allowedDAOs[:],
	}

	pid := proposals.SetProposal(p)
	dao.PostCreateProposal(0, cur, r, pid)

	chain.Emit("ProposalCreated",
		"id", strconv.FormatInt(int64(pid), 10),
	)

	return pid, nil
}

func MustVoteOnProposal(cur realm, r VoteRequest) {
	if err := VoteOnProposal(cur, r); err != nil {
		panic(err.Error())
	}
}

// VoteOnProposal sends a vote to the actual govDAO implementation.
// If the voter cannot vote the specified proposal, this method will return an error
// with the explanation of why.
func VoteOnProposal(cur realm, r VoteRequest) error {
	if dao == nil {
		return errors.New("DAO not initialized")
	}
	return dao.VoteOnProposal(0, cur, r)
}

// MustVoteOnProposalSimple is like MustVoteOnProposal but intended to be used through gnokey with basic types.
func MustVoteOnProposalSimple(cur realm, pid int64, option string) {
	MustVoteOnProposal(cur, VoteRequest{
		Option:     VoteOption(option),
		ProposalID: ProposalID(pid),
	})
}

func MustGetProposal(pid ProposalID) *Proposal {
	p, err := GetProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	return p
}

// GetProposal gets created proposal by its ID. Non-crossing pure read:
// looks up the proposal in this realm's package var. Callable directly
// from any realm without cross-call syntax.
func GetProposal(pid ProposalID) (*Proposal, error) {
	if dao == nil {
		return nil, errors.New("DAO not initialized")
	}
	prop := proposals.GetProposal(pid)
	if prop == nil {
		return nil, errors.New(ufmt.Sprintf("Proposal %v does not exist.", int64(pid)))
	}
	return prop, nil
}

// UpdateImpl is a method intended to be used on a proposal.
// This method will update the current govDAO implementation
// to a new one. AllowedDAOs are a list of realms that can
// call this method, in case the new DAO implementation had
// a breaking bug. Any value set as nil will be ignored.
// If AllowedDAOs field is not set correctly, the actual DAO
// implementation wont be able to execute new Proposals!
func UpdateImpl(cur realm, r UpdateRequest) {
	gRealm := cur.Previous().PkgPath()

	if !InAllowedDAOs(gRealm) {
		panic("permission denied for prev realm: " + gRealm)
	}

	if r.AllowedDAOs != nil {
		allowedDAOs = r.AllowedDAOs
	}

	if r.DAO != nil {
		dao = r.DAO
	}
}

func AllowedDAOs() []string {
	dup := make([]string, len(allowedDAOs))
	copy(dup, allowedDAOs)
	return dup
}

func InAllowedDAOs(pkg string) bool {
	if len(allowedDAOs) == 0 {
		return true // corner case for initialization
	}
	for _, d := range allowedDAOs {
		if pkg == d {
			return true
		}
	}
	return false
}

func executeProposal(cur realm, pid ProposalID, execErrorRejects bool) bool {
	if dao == nil {
		return false
	}
	execute, err := dao.PreExecuteProposal(0, cur, pid)
	if err != nil {
		panic(err.Error())
	}

	if !execute {
		return false
	}
	prop, err := GetProposal(pid)
	if err != nil {
		panic(err.Error())
	}

	err = dao.ExecuteProposal(0, cur, pid, prop.executor)
	if err != nil {
		if execErrorRejects {
			return false
		}

		panic(err.Error())
	}
	return true
}
