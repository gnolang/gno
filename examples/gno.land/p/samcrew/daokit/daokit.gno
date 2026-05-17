package daokit

import (
	"time"

	"gno.land/p/samcrew/daocond"
)

// DAO is the internal dispatch interface for a DAO. All methods are
// address-typed (no rlm crosses the interface boundary) so a malicious
// implementer cannot capture a cur from a consumer's wrapper. This is the
// Authority-pattern: the trusted rlm-validating wrapper (e.g.
// *basedao.DAOWrapper) is the only public surface; this interface is
// dispatch-only.
//
// CONTRACT: NEVER store this interface as a public package var, NEVER
// accept it as a public function parameter, and NEVER return it from a
// public constructor. Public APIs should expose a concrete wrapper type
// (such as *basedao.DAOWrapper) whose rlm-typed methods derive the
// caller from cur.Previous().Address() under an IsCurrent() guard.
//
// Sealing via an unexported marker method would NOT help — embedding
// bypasses the seal; see p/test/seal/filetests/z_seal_*_filetest.gno.
type DAO interface {
	Propose(proposer address, req ProposalRequest) uint64
	Execute(executor address, id uint64)
	Vote(voter address, id uint64, vote daocond.Vote)
}

type Core struct {
	Resources *ResourcesStore
	Proposals *ProposalsStore
}

func NewCore() *Core {
	return &Core{
		Resources: NewResourcesStore(),
		Proposals: NewProposalsStore(),
	}
}

func (d *Core) Vote(voterID string, proposalID uint64, vote daocond.Vote) {
	proposal := d.Proposals.GetProposal(proposalID)
	if proposal == nil {
		panic("proposal not found")
	}

	if proposal.Status != ProposalStatusOpen {
		panic("proposal is not open")
	}

	proposal.Ballot.Vote(voterID, daocond.Vote(vote))
}

// Execute runs the proposal's registered action handler via the generic
// daokit.ActionHandler.Execute path. Equivalent to
// ExecuteWithDispatch(proposalID, nil).
func (d *Core) Execute(proposalID uint64) {
	d.ExecuteWithDispatch(proposalID, nil)
}

// ExecuteWithDispatch validates the proposal (existence, Open status,
// Condition.Eval, status → Passed) and then either runs the registered
// Handler or yields to the caller-supplied dispatch override. If
// `dispatch(action)` returns true the generic Handler.Execute call is
// skipped — the caller is asserting that it dispatched the action via
// its own path. Used by basedao to thread a live `rlm` into the
// EditProfile dispatcher without putting `rlm` on the public
// daokit.ActionHandler interface (Class-1a defense — see
// docs/resources/gno-security.md).
//
// The dispatch callback receives only the public `Action` value; no
// `rlm` flows through this function's signature. rlm-bearing dispatch
// is the caller's responsibility, captured in their closure.
//
// CAVEAT (documented contract-violation symptom): if the registered
// generic Handler.Execute panics — e.g. basedao's editProfileStubHandler
// when EditProfile is reached via the legacy DAOPrivate.Execute →
// Core.Execute path — the proposal's status has already moved
// Open → Passed (by UpdateStatus above), but Status = Executed is never
// set. The proposal becomes stuck in Passed. This is only reachable
// when a caller violates the "DAOPrivate is internal-only" contract
// (basedao.gno DAOWrapper.Private() doc); production callers go through
// DAOWrapper.Execute → DAOPrivate.executeWithRlm → here with a non-nil
// dispatch that intercepts EditProfile before the stub is reached.
func (d *Core) ExecuteWithDispatch(proposalID uint64, dispatch func(Action) bool) {
	proposal := d.Proposals.GetProposal(proposalID)
	if proposal == nil {
		panic("proposal not found")
	}

	if proposal.Status != ProposalStatusOpen {
		panic("proposal is not open")
	}

	if !proposal.Condition.Eval(proposal.Ballot) {
		panic("proposal condition is not met")
	}

	proposal.UpdateStatus()
	if proposal.Status != ProposalStatusPassed {
		panic("proposal does not meet the condition(s) or is already closed/executed")
	}

	if dispatch == nil || !dispatch(proposal.Action) {
		d.Resources.Get(proposal.Action.Type()).Handler.Execute(proposal.Action)
	}
	proposal.Status = ProposalStatusExecuted
	proposal.ExecutedAt = time.Now()
}

func (d *Core) Propose(proposerID string, req ProposalRequest) uint64 {
	actionType := req.Action.Type()

	resource := d.Resources.Get(actionType)
	if resource == nil {
		panic("action type is not registered as a resource")
	}

	prop := d.Proposals.newProposal(proposerID, req, resource.Condition)
	return uint64(prop.ID)
}

func (d *Core) ResourcesCount() int {
	return d.Resources.Tree.Size()
}
