package commondao

import (
	"time"

	"gno.land/p/moul/addrset"
)

// ReadonlyCommonDAO is a read only view of a CommonDAO. It exposes only
// read side methods, holds the *CommonDAO in an unexported field, and every
// reachable value is itself readonly or a copy — so cross-realm holders
// cannot mutate the DAO through it. Views are live handles, not snapshots.
//
// This is the only safe handle to a DAO across a realm boundary: hosting
// realms must never return the *CommonDAO itself.
type ReadonlyCommonDAO struct {
	dao *CommonDAO
}

// Readonly returns a read only view of the DAO.
func (dao *CommonDAO) Readonly() ReadonlyCommonDAO {
	return ReadonlyCommonDAO{dao}
}

// ID returns DAO's unique identifier.
func (r ReadonlyCommonDAO) ID() uint64 {
	return r.dao.id
}

// Name returns DAO's name.
func (r ReadonlyCommonDAO) Name() string {
	return r.dao.name
}

// Description returns DAO's description.
func (r ReadonlyCommonDAO) Description() string {
	return r.dao.description
}

// Address returns the DAO's treasury address, or empty when unset.
func (r ReadonlyCommonDAO) Address() address {
	return r.dao.addr
}

// IsDeleted returns true when DAO has been soft deleted.
func (r ReadonlyCommonDAO) IsDeleted() bool {
	return r.dao.deleted
}

// IsTreasuryFrozen checks if the DAO's treasury is frozen.
func (r ReadonlyCommonDAO) IsTreasuryFrozen() bool {
	return r.dao.treasuryFrozen
}

// Council returns a read only view of the DAO council.
func (r ReadonlyCommonDAO) Council() *addrset.ReadonlySet {
	return r.dao.council.Readonly()
}

// Parent returns a read only view of the parent DAO when there is one.
func (r ReadonlyCommonDAO) Parent() (_ ReadonlyCommonDAO, found bool) {
	if r.dao.parent == nil {
		return ReadonlyCommonDAO{}, false
	}
	return r.dao.parent.Readonly(), true
}

// ChildrenCount returns the number of direct children DAOs.
func (r ReadonlyCommonDAO) ChildrenCount() int {
	return r.dao.children.Len()
}

// IterateChildren iterates the direct children DAOs.
// The callback can return true to stop iteration.
func (r ReadonlyCommonDAO) IterateChildren(fn func(ReadonlyCommonDAO) bool) (stopped bool) {
	r.dao.children.ForEach(func(_ int, v any) bool {
		stopped = fn(v.(*CommonDAO).Readonly())
		return stopped
	})
	return stopped
}

// GetProposal returns a read only view of a proposal when it exists.
func (r ReadonlyCommonDAO) GetProposal(proposalID uint64) (_ ReadonlyProposal, found bool) {
	p := r.dao.GetProposal(proposalID)
	if p == nil {
		return ReadonlyProposal{}, false
	}
	return p.Readonly(), true
}

// ActiveProposalsSize returns the number of active proposals.
func (r ReadonlyCommonDAO) ActiveProposalsSize() int {
	return r.dao.activeProposals.Size()
}

// FinishedProposalsSize returns the number of finished proposals.
func (r ReadonlyCommonDAO) FinishedProposalsSize() int {
	return r.dao.finishedProposals.Size()
}

// IterateActiveProposals iterates read only views of the active proposals.
// The callback can return true to stop iteration.
func (r ReadonlyCommonDAO) IterateActiveProposals(offset, count int, reverse bool, fn func(ReadonlyProposal) bool) bool {
	return r.dao.activeProposals.Iterate(offset, count, reverse, func(p *Proposal) bool {
		return fn(p.Readonly())
	})
}

// IterateFinishedProposals iterates read only views of the finished proposals.
// The callback can return true to stop iteration.
func (r ReadonlyCommonDAO) IterateFinishedProposals(offset, count int, reverse bool, fn func(ReadonlyProposal) bool) bool {
	return r.dao.finishedProposals.Iterate(offset, count, reverse, func(p *Proposal) bool {
		return fn(p.Readonly())
	})
}

// ReadonlyProposal is a read only view of a Proposal. Proposal content is
// exposed flattened (Title, Body): the view never exposes the underlying
// ProposalDefinition, whose Executor would otherwise be callable by any
// holder with the hosting realm's authority.
type ReadonlyProposal struct {
	p *Proposal
}

// Readonly returns a read only view of the proposal.
func (p *Proposal) Readonly() ReadonlyProposal {
	return ReadonlyProposal{p}
}

// ID returns the unique proposal identifier.
func (r ReadonlyProposal) ID() uint64 {
	return r.p.id
}

// Status returns the current proposal status.
func (r ReadonlyProposal) Status() ProposalStatus {
	return r.p.status
}

// StatusReason returns an optional reason that lead to the current proposal status.
func (r ReadonlyProposal) StatusReason() string {
	return r.p.statusReason
}

// Creator returns the address of the account that created the proposal.
func (r ReadonlyProposal) Creator() address {
	return r.p.creator
}

// CreatedAt returns the time that proposal was created.
func (r ReadonlyProposal) CreatedAt() time.Time {
	return r.p.createdAt
}

// VotingDeadline returns the deadline after which no more votes are allowed.
func (r ReadonlyProposal) VotingDeadline() time.Time {
	return r.p.votingDeadline
}

// Title returns the proposal definition's title.
func (r ReadonlyProposal) Title() string {
	return r.p.definition.Title()
}

// Body returns the proposal definition's body.
func (r ReadonlyProposal) Body() string {
	return r.p.definition.Body()
}

// VoteChoices returns the list of vote choices allowed for the proposal.
func (r ReadonlyProposal) VoteChoices() []VoteChoice {
	return r.p.VoteChoices()
}

// VotingRecord returns a read only record with the submitted votes.
func (r ReadonlyProposal) VotingRecord() ReadonlyVotingRecord {
	return r.p.record.Readonly()
}

// Electorate returns a read only view of the proposal's electorate.
func (r ReadonlyProposal) Electorate() *addrset.ReadonlySet {
	return r.p.Electorate()
}
