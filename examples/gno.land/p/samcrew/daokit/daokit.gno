package daokit

import (
	"time"

	"gno.land/p/samcrew/daocond"
)

type DAO interface {
	// Creates a new proposal and returns its ID.
	Propose(req ProposalRequest) uint64
	// Casts a vote on a specific proposal.
	Vote(id uint64, vote daocond.Vote)
	// Executes a proposal if it meets the required conditions.
	Execute(id uint64)

	// Generates a web interface representation for the given path.
	Render(path string) string

	// Gets an extension by path, returns nil if not found.
	Extension(path string) Extension
	// Lists all available extensions.
	ExtensionsList() ExtensionsList
}

// Function type for updating DAO implementation during governance upgrades.
type SetImplemFn = func(implem DAO)

// Creates, votes yes, and immediately executes a proposal in one operation.
// Useful when you have enough permission to execute an action directly.
// Examples: migrations, adding members
func InstantExecute(d DAO, req ProposalRequest) uint64 {
	id := d.Propose(req)
	d.Vote(id, daocond.VoteYes)
	d.Execute(id)
	return id
}

// Manages the essential components of a DAO: resources, proposals, and extensions.
type Core struct {
	Resources  *ResourcesStore  // Available actions and their conditions
	Proposals  *ProposalsStore  // All proposals and their voting state
	Extensions *ExtensionsStore // Pluggable functionality modules
}

// Creates a new DAO core with empty stores.
func NewCore() *Core {
	return &Core{
		Resources:  NewResourcesStore(),
		Proposals:  NewProposalsStore(),
		Extensions: &ExtensionsStore{},
	}
}

// Records a vote for a specific proposal from a voter.
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

// Executes a proposal if it meets the required voting conditions.
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

// Creates a new proposal for voting and returns its ID.
func (d *Core) Propose(proposerID string, req ProposalRequest) uint64 {
	actionType := req.Action.Type()

	resource := d.Resources.Get(actionType)
	if resource == nil {
		panic("action type is not registered as a resource")
	}

	prop := d.Proposals.newProposal(proposerID, req, resource.Condition)
	return uint64(prop.ID)
}
