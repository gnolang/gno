package daokit

import (
	"time"

	"gno.land/p/samcrew/daocond"
)

// DAO is the public interface for a DAO. All mutating methods take
// (_ int, rlm realm) so the implementation can derive the caller's
// identity from rlm.Previous().Address() under an rlm.IsCurrent()
// guard — unforgeable, unlike the pre-migration runtime walk path
// where a non-crossing wrapper could spoof the principal.
type DAO interface {
	Propose(_ int, rlm realm, req ProposalRequest) uint64
	Execute(_ int, rlm realm, id uint64)
	Vote(_ int, rlm realm, id uint64, vote daocond.Vote)
}

func InstantExecute(_ int, rlm realm, d DAO, req ProposalRequest) uint64 {
	id := d.Propose(0, rlm, req)
	d.Vote(0, rlm, id, daocond.VoteYes)
	d.Execute(0, rlm, id)
	return id
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
