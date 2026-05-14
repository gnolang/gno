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

func (d *Core) Execute(proposalID uint64) {
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

	d.Resources.Get(proposal.Action.Type()).Handler.Execute(proposal.Action)
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
