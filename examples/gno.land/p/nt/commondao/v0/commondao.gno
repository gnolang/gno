package commondao

import (
	"errors"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/bptree/v0/list"
	"gno.land/p/nt/seqid/v0"
)

// DefaultMaxActiveProposals is the default cap for simultaneously active
// proposals per DAO. Every active proposal stores a council snapshot, so
// the cap bounds storage; it can be changed with WithMaxActiveProposals.
const DefaultMaxActiveProposals = 32

var (
	ErrCouncilUpdateOverlap  = errors.New("council update adds and removes the same address")
	ErrDAOIsDeleted          = errors.New("DAO is deleted")
	ErrEmptyCouncil          = errors.New("council update would remove every council member")
	ErrExecutionNotAllowed   = errors.New("proposal must be active or passed to be executed")
	ErrInvalidVoteChoice     = errors.New("invalid vote choice")
	ErrMaxActiveProposals    = errors.New("max number of active proposals reached")
	ErrMaxCapExemptProposals = errors.New("creator already has an active cap exempt proposal")
	ErrNotElectorateMember   = errors.New("account is not a member of the proposal's electorate")
	ErrOverflow              = errors.New("next ID overflows uint64")
	ErrProposalNotFound      = errors.New("proposal not found")
	ErrVotingDeadlineNotMet  = errors.New("voting deadline not met")
	ErrVotingDeadlinePassed  = errors.New("voting deadline has passed")
	ErrWithdrawalNotAllowed  = errors.New("withdrawal not allowed for proposals with votes")
)

// CommonDAO defines a DAO.
//
// # Security
//
// A *CommonDAO is a mutable handle: its exported mutators (UpdateCouncil,
// Dissolve, Propose, Vote, Execute, Withdraw) are meant for the realm
// that owns the DAO. Three rules apply at realm boundaries:
//
//  1. Do not ACCEPT a *CommonDAO from an external/untrusted caller.
//  2. Do not RETURN a *CommonDAO from any function callable by untrusted
//     realms — return dao.Readonly() (a ReadonlyCommonDAO view) instead.
//  3. Do not TRUST a readonly view received from an untrusted caller: it
//     is a live handle over the sender's data.
type CommonDAO struct {
	id                 uint64
	name               string
	description        string
	addr               address // derived treasury address, empty when unset
	parent             *CommonDAO
	children           list.IList
	council            *addrset.Set
	genID              seqid.ID
	activeProposals    *proposalStorage
	finishedProposals  *proposalStorage
	deleted            bool // Soft delete
	treasuryFrozen     bool
	maxActiveProposals int
}

// New creates a new common DAO.
func New(options ...Option) *CommonDAO {
	dao := &CommonDAO{
		children:           &list.List{},
		council:            &addrset.Set{},
		activeProposals:    newProposalStorage(),
		finishedProposals:  newProposalStorage(),
		maxActiveProposals: DefaultMaxActiveProposals,
	}
	for _, apply := range options {
		apply(dao)
	}
	return dao
}

// ID returns DAO's unique identifier.
func (dao CommonDAO) ID() uint64 {
	return dao.id
}

// Name returns DAO's name.
func (dao CommonDAO) Name() string {
	return dao.name
}

// Description returns DAO's description.
func (dao CommonDAO) Description() string {
	return dao.description
}

// Address returns the DAO's treasury address, assigned at creation with
// WithAddress. The package never derives or uses the address itself:
// hosting realms derive it (e.g. from a realm sub-identity) and operate
// its funds through their own banker. Empty when unset.
func (dao CommonDAO) Address() address {
	return dao.addr
}

// IsTreasuryFrozen checks if the DAO's treasury is frozen. The package
// stores the flag only; hosting realms enforce it when moving funds.
func (dao CommonDAO) IsTreasuryFrozen() bool {
	return dao.treasuryFrozen
}

// SetTreasuryFrozen freezes or unfreezes the DAO's treasury.
func (dao *CommonDAO) SetTreasuryFrozen(frozen bool) {
	dao.treasuryFrozen = frozen
}

// Parent returns the parent DAO.
// Null can be returned when DAO has no parent assigned.
func (dao CommonDAO) Parent() *CommonDAO {
	return dao.parent
}

// ChildrenCount returns the number of direct children DAOs.
func (dao CommonDAO) ChildrenCount() int {
	return dao.children.Len()
}

// IterateChildren iterates the direct children DAOs.
func (dao CommonDAO) IterateChildren(fn func(*CommonDAO) bool) (stopped bool) {
	dao.children.ForEach(func(_ int, v any) bool {
		return fn(v.(*CommonDAO))
	})
	return
}

// TopParent returns the topmost parent DAO.
// The top parent is the root of the DAO tree.
func (dao *CommonDAO) TopParent() *CommonDAO {
	parent := dao.Parent()
	if parent != nil {
		return parent.TopParent()
	}
	return dao
}

// Council returns a read only view of the DAO council.
//
// The council is the set of addresses entitled to vote. It changes only
// through UpdateCouncil (normally called by a council update proposal
// executor) or constructor options.
func (dao CommonDAO) Council() *addrset.ReadonlySet {
	return dao.council.Readonly()
}

// UpdateCouncil adds and removes council members as idempotent set
// operations: adding an existing member or removing an absent one is a
// no-op, so concurrently passed council updates merge deterministically in
// execution order, and a full council replacement in a single update is
// legal.
//
// The final set is (council ∪ add) \ remove. An update that adds and
// removes the same address is rejected, and an update whose final set
// would empty a non-empty council returns ErrEmptyCouncil: executors must
// propagate the error (failing the proposal) instead of panicking, which
// would revert the transaction and leave the proposal stuck.
func (dao *CommonDAO) UpdateCouncil(add, remove []address) error {
	for _, a := range add {
		for _, r := range remove {
			if a == r {
				return ErrCouncilUpdateOverlap
			}
		}
	}

	// The final set can only be empty when nothing is added: overlap is
	// rejected above, so any added address survives its own update.
	if dao.council.Size() > 0 && len(add) == 0 {
		empty := true
		dao.council.IterateByOffset(0, dao.council.Size(), func(member address) bool {
			for _, r := range remove {
				if r == member {
					return false // removed: keep looking for a survivor
				}
			}
			empty = false
			return true
		})
		if empty {
			return ErrEmptyCouncil
		}
	}

	for _, a := range add {
		dao.council.Add(a)
	}
	for _, r := range remove {
		dao.council.Remove(r)
	}
	return nil
}

// ActiveProposalsSize returns the number of active proposals, including
// early passed proposals that were not executed yet.
func (dao CommonDAO) ActiveProposalsSize() int {
	return dao.activeProposals.Size()
}

// IterateActiveProposals iterates active proposals ordered by ID.
func (dao CommonDAO) IterateActiveProposals(offset, count int, reverse bool, fn func(*Proposal) bool) bool {
	return dao.activeProposals.Iterate(offset, count, reverse, fn)
}

// FinishedProposalsSize returns the number of finished proposals.
func (dao CommonDAO) FinishedProposalsSize() int {
	return dao.finishedProposals.Size()
}

// IterateFinishedProposals iterates finished proposals ordered by ID.
func (dao CommonDAO) IterateFinishedProposals(offset, count int, reverse bool, fn func(*Proposal) bool) bool {
	return dao.finishedProposals.Iterate(offset, count, reverse, fn)
}

// IsDeleted returns true when DAO has been soft deleted.
func (dao CommonDAO) IsDeleted() bool {
	return dao.deleted
}

// MaxActiveProposals returns the cap for simultaneously active proposals.
func (dao CommonDAO) MaxActiveProposals() int {
	return dao.maxActiveProposals
}

// SetMaxActiveProposals changes the cap for simultaneously active
// proposals. Values below one are ignored: a DAO must always be able to
// propose.
func (dao *CommonDAO) SetMaxActiveProposals(max int) {
	if max >= 1 {
		dao.maxActiveProposals = max
	}
}

// Propose creates a new DAO proposal.
//
// The proposal's electorate is the council snapshot taken now: members
// added later vote on the next proposal; members removed later remain in
// the electorate, where their silence counts against passage.
//
// The number of simultaneously active proposals is capped. Definitions
// implementing CapExempt (e.g. council updates, which must never be
// blockable by a full cap) are exempt but bounded to one active proposal
// per creator.
func (dao *CommonDAO) Propose(creator address, d ProposalDefinition) (*Proposal, error) {
	if dao.deleted {
		return nil, ErrDAOIsDeleted
	}

	if d == nil {
		return nil, ErrProposalDefinitionRequired
	}

	// Every definition is decided by the default Council voting rules.
	if _, ok := d.(DefaultTallier); !ok {
		return nil, ErrMissingTallier
	}

	if _, exempt := d.(CapExempt); exempt {
		var found bool
		dao.activeProposals.Iterate(0, dao.activeProposals.Size(), false, func(p *Proposal) bool {
			if _, ok := p.definition.(CapExempt); ok && p.creator == creator {
				found = true
				return true
			}
			return false
		})
		if found {
			return nil, ErrMaxCapExemptProposals
		}
	} else if dao.activeProposals.Size() >= dao.maxActiveProposals {
		return nil, ErrMaxActiveProposals
	}

	id, ok := dao.genID.TryNext()
	if !ok {
		return nil, ErrOverflow
	}

	p, err := newProposal(uint64(id), creator, d)
	if err != nil {
		return nil, err
	}

	// Snapshot the current council as the proposal's electorate
	dao.council.IterateByOffset(0, dao.council.Size(), func(member address) bool {
		p.electorate.Add(member)
		return false
	})

	dao.activeProposals.Add(p)
	return p, nil
}

// GetProposal returns a proposal or nil when proposal is not found.
func (dao CommonDAO) GetProposal(proposalID uint64) *Proposal {
	p := dao.activeProposals.Get(proposalID)
	if p != nil {
		return p
	}
	return dao.finishedProposals.Get(proposalID)
}

// Withdraw withdraws a proposal that has no votes.
// Only active proposals without votes can be withdrawn, and once
// withdrawn they are considered finished.
func (dao *CommonDAO) Withdraw(proposalID uint64) error {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	if p.status != StatusActive {
		return ErrStatusIsNotActive
	}

	if p.record.Size() > 0 {
		return ErrWithdrawalNotAllowed
	}

	p.status = StatusWithdrawn
	dao.activeProposals.Remove(p.id)
	dao.finishedProposals.Add(p)
	return nil
}

// Vote submits a new vote for a proposal.
//
// Votes are only allowed to members of the proposal's electorate while the
// proposal is active and within the voting period. A member may change
// their vote by voting again.
//
// Proposals decided by the default Council rules (DefaultTallier
// definitions) are re-evaluated after every recorded vote: a supermajority
// (or the definition's threshold) of YES decides the proposal immediately,
// and a simple majority of NO dismisses it immediately.
func (dao *CommonDAO) Vote(member address, proposalID uint64, c VoteChoice, reason string) error {
	if dao.deleted {
		return ErrDAOIsDeleted
	}

	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	if p.status != StatusActive {
		return ErrStatusIsNotActive
	}

	if !p.electorate.Has(member) {
		return ErrNotElectorateMember
	}

	if p.HasVotingDeadlinePassed() {
		return ErrVotingDeadlinePassed
	}

	if !p.IsVoteChoiceValid(c) {
		return ErrInvalidVoteChoice
	}

	p.record.AddVote(Vote{
		Address: member,
		Choice:  c,
		Reason:  reason,
	})

	// Early termination: default rule proposals are decided the moment
	// the outcome is mathematically settled. A passed proposal stays in
	// the active storage until executed; a dismissed one is finished.
	if dt, ok := p.definition.(DefaultTallier); ok {
		switch TallyDefault(p.record.Readonly(), p.Electorate(), dt.Threshold()) {
		case OutcomePassed:
			p.status = StatusPassed
		case OutcomeDismissed:
			dao.dismiss(p)
		}
	}
	return nil
}

// Execute executes a proposal.
//
// Proposals that already passed (decided early by the default Council
// rules) execute immediately. Active proposals are tallied once their
// voting deadline passes and are dismissed unless passed.
//
// rlm is the realm authority threaded into the executor via cross(rlm).
// Execute itself is not a crossing function (rlm sits in a non-first
// parameter slot) because /p/ production code cannot declare crossing
// functions — callers pass `cur` directly as the rlm argument.
func (dao *CommonDAO) Execute(proposalID uint64, rlm realm) error {
	if dao.deleted {
		return ErrDAOIsDeleted
	}

	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	switch p.status {
	case StatusPassed:
		// Decided early: execute now, before the voting deadline
	case StatusActive:
		if !p.HasVotingDeadlinePassed() {
			return ErrVotingDeadlineNotMet
		}
	default:
		return ErrExecutionNotAllowed
	}

	// The proposal leaves active storage before any definition code
	// (Validate, Tally, the executor) runs, so a re-entrant Execute
	// call cannot run it twice.
	dao.activeProposals.Remove(p.id)

	// IMPORTANT, from this point on, any error is going to result
	// in a proposal failure and execute will succeed.

	// Validate proposal before execution
	err := p.Validate()

	// Tally proposals that are still active after their deadline;
	// undecided proposals are dismissed
	if err == nil && p.status == StatusActive {
		if dt, ok := p.definition.(DefaultTallier); ok {
			if TallyDefault(p.record.Readonly(), p.Electorate(), dt.Threshold()) == OutcomePassed {
				p.status = StatusPassed
			}
		}

		if p.status != StatusPassed {
			p.status = StatusDismissed
			dao.finishedProposals.Add(p)
			return nil
		}
	}

	// Execute proposal only if it's executable
	if err == nil {
		if e, ok := p.Definition().(Executable); ok {
			if fn := e.Executor(); fn != nil {
				err = fn(cross(rlm))
			}
		}
	}

	// Proposal fails if there is any error during validation and execution process
	if err != nil {
		p.status = StatusFailed
		p.statusReason = err.Error()
	} else {
		p.status = StatusExecuted
		p.statusReason = ""
	}

	// Whichever the outcome of the validation, tallying
	// and execution consider the proposal finished.
	dao.finishedProposals.Add(p)
	return nil
}

// dismiss finishes a proposal as dismissed.
func (dao *CommonDAO) dismiss(p *Proposal) {
	p.status = StatusDismissed
	dao.activeProposals.Remove(p.id)
	dao.finishedProposals.Add(p)
}

// Dissolve soft deletes the DAO after dismissing every in-flight proposal
// (both still-active and passed-but-unexecuted ones). Dissolution is
// terminal: a deleted DAO rejects proposals, votes and executions, so
// nothing may remain pending.
func (dao *CommonDAO) Dissolve(reason string) {
	var pending []*Proposal
	dao.activeProposals.Iterate(0, dao.activeProposals.Size(), false, func(p *Proposal) bool {
		pending = append(pending, p)
		return false
	})

	for _, p := range pending {
		p.statusReason = reason
		dao.dismiss(p)
	}
	dao.deleted = true
}
