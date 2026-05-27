// Package daokit is a reference DAO toolkit (propose / vote / execute,
// resource registration, action handlers) used internally by
// p/samcrew/basedao and the r/samcrew/daodemo example realm.
//
// ─── STATUS: EXPERIMENTAL / INSECURE / INCOMPLETE / COMPLICATED ───
//
// Do not use as the basis for a production DAO without significant
// hardening. The gnoland chain's actual governance is r/gov/dao/v3
// (independent code path; not built on daokit).
//
// Known issues (non-exhaustive):
//
//   - Re-entrancy: Core.ExecuteWithDispatch evaluates the
//     caller-supplied daocond.Condition.Eval BEFORE mutating proposal
//     status. A hostile Condition can re-enter Propose/Vote/Execute
//     mid-flight on the same DAO. The SECURITY XXX in actions.gno
//     (lines 17-26) proposes a per-proposal in-flight flag — not
//     implemented.
//
//   - Status state machine is incomplete: there is no Failed or
//     Closed state. ExecuteWithDispatch moves Status Open→Passed
//     BEFORE invoking the handler, then sets Executed only on
//     success. A handler panic leaves the proposal stuck in Passed
//     with no path to retry, fail, or cancel.
//
//   - ResourcesStore.Set silently overwrites entries by Action kind.
//     The Handler is re-fetched at execute time, so a malicious
//     owner can swap the handler between propose and execute. The
//     captured Condition is immune (resolved at propose time) but
//     Handler is not.
//
//   - Core.Vote does not bound the Vote payload size. Anyone
//     reaching Core directly (e.g. via *DAOPrivate.Core in basedao)
//     can grief storage. basedao.DAOPrivate.Vote bounds at 16 bytes;
//     the underlying Core does not.
//
//   - daokit.ActionHandler and daocond.Condition are open interfaces.
//     Hostile implementations registered via Resources.Set run with
//     the DAO's authority — same trust level as code inside the DAO
//     package itself. This is documented in actions.gno but not
//     enforced.
//
//   - Stale comment at actions.gno:102-108 references a removed
//     InstantExecute free function that no longer exists.
//
// Architectural complexity notes:
//
//   - The Action / ActionHandler / Resources triangle requires three
//     coordinated registrations per action kind (Action type,
//     Handler closure, Resource entry with Condition).
//   - basedao layers a DAOWrapper / DAOPrivate split on top.
//     DAOWrapper.ExecuteWithDispatch routes actions that need
//     cross-realm authority through an rlm-aware handler before
//     falling back to the address-typed Core.Execute path. Reading
//     the execute path requires tracing through three packages.
//
// Use at your own risk. If you adopt this package, fork it and
// harden the items above before exposing it to untrusted input.
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
