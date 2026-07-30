package commondao

import (
	"errors"

	"gno.land/p/moul/addrset"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/bptree/v0/list"
	"gno.land/p/nt/seqid/v0"
)

// DefaultMaxActiveProposals is the default cap for simultaneously active
// proposals per DAO. Every active proposal stores a council snapshot, so
// the cap bounds storage. It is applied at construction; a hosting realm
// may override it per DAO via SetMaxActiveProposals (the reference realm
// keeps this default).
const DefaultMaxActiveProposals = 32

var (
	ErrCannotDeregisterRegisterKind = errors.New("the register-kind proposal kind cannot be deregistered")

	ErrCouncilUpdateOverlap  = errors.New("council update adds and removes the same address")
	ErrDAOIsDeleted          = errors.New("DAO is deleted")
	ErrEmptyCouncil          = errors.New("council update would remove every council member")
	ErrExecutionNotAllowed   = errors.New("proposal must be active or passed to be executed")
	ErrInvalidVoteChoice     = errors.New("invalid vote choice")
	ErrMaxActiveProposals    = errors.New("max number of active proposals reached")
	ErrMaxCapExemptProposals = errors.New("creator already has an active cap exempt proposal")
	ErrNotElectorateMember   = errors.New("account is not a member of the proposal's electorate")
	ErrOverflow              = errors.New("next ID overflows uint64")
	ErrProposalKindExists    = errors.New("proposal kind already registered")
	ErrProposalKindNotFound  = errors.New("proposal kind not found")
	ErrProposalKindRequired  = errors.New("proposal kind is required")
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
// Dissolve, Propose, Vote, Execute, Withdraw, SetTreasuryFrozen,
// SetMaxActiveProposals, RegisterKind, DeregisterKind) are meant for the
// realm that owns the DAO.
// Three rules apply at realm boundaries:
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
	purpose            string
	addr               address // derived treasury address, empty when unset
	parent             *CommonDAO
	children           list.IList
	council            *addrset.Set
	genID              seqid.ID
	kinds              *bptree.BPTree // proposal kind name -> ProposalKind
	activeProposals    *proposalStorage
	finishedProposals  *proposalStorage
	deleted            bool // Soft delete
	treasuryFrozen     bool
	maxActiveProposals int
	proposing          bool // re-entrancy latch around a kind's New in Propose
	executing          bool // re-entrancy latch around Execute
}

// New creates a new common DAO.
func New(options ...Option) *CommonDAO {
	dao := &CommonDAO{
		children:           &list.List{},
		council:            &addrset.Set{},
		kinds:              bptree.NewBPTree32(),
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

// Purpose returns the DAO's purpose. Together with the description it
// forms the DAO's Charter (docs/CONSTITUTION.md :1485).
func (dao CommonDAO) Purpose() string {
	return dao.purpose
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
		stopped = fn(v.(*CommonDAO))
		return stopped
	})
	return stopped
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

// RegisterKind registers a proposal kind by its name.
//
// Registered kinds are the only way to create proposals: Propose looks
// kinds up by name and calls their New factory, so a proposal type is
// proposable iff its kind is registered. Like SetTreasuryFrozen, this
// mutator is meant for the realm that owns the DAO (typically called at
// DAO creation and by governance proposal executors).
func (dao *CommonDAO) RegisterKind(k ProposalKind) error {
	if k == nil || k.Name() == "" {
		return ErrProposalKindRequired
	}
	if dao.kinds.Has(k.Name()) {
		return ErrProposalKindExists
	}
	dao.kinds.Set(k.Name(), k)
	return nil
}

// DeregisterKind removes a proposal kind by name.
//
// Deregistering only blocks new proposals: the registry is read at
// Propose time only, so in-flight proposals of the kind keep their
// frozen definition and still vote and execute.
//
// The reserved register-kind name cannot be deregistered: removing it
// would brick governance-driven registration (the only way to add kinds
// back once the owning realm no longer holds the handle), so the call
// returns ErrCannotDeregisterRegisterKind instead.
func (dao *CommonDAO) DeregisterKind(name string) error {
	if name == registerKindName {
		return ErrCannotDeregisterRegisterKind
	}
	if _, removed := dao.kinds.Remove(name); !removed {
		return ErrProposalKindNotFound
	}
	return nil
}

// HasKind checks if a proposal kind is registered.
func (dao CommonDAO) HasKind(name string) bool {
	return dao.kinds.Has(name)
}

// KindNames returns the names of the registered proposal kinds, sorted.
func (dao CommonDAO) KindNames() []string {
	names := make([]string, 0, dao.kinds.Size())
	dao.kinds.IterateByOffset(0, dao.kinds.Size(), func(name string, _ any) bool {
		names = append(names, name)
		return false
	})
	return names
}

// Propose creates a new DAO proposal.
//
// Proposals are created through registered proposal kinds: the kind is
// looked up by name in the DAO's registry and its New factory builds the
// proposal definition from args. The registry is read only here and the
// definition is frozen once the proposal is created, so deregistering a
// kind later never touches in-flight proposals.
//
// The proposal's electorate is the council snapshot taken now: members
// added later vote on the next proposal; members removed later remain in
// the electorate, where their silence counts against passage.
//
// The number of simultaneously active proposals is capped. Definitions
// implementing CapExempt (e.g. council updates, which must never be
// blockable by a full cap) are exempt but bounded to one active proposal
// per creator.
func (dao *CommonDAO) Propose(creator address, kind string, args any) (*Proposal, error) {
	if dao.deleted {
		return nil, ErrDAOIsDeleted
	}

	v := dao.kinds.Get(kind)
	if v == nil {
		return nil, ErrProposalKindNotFound
	}

	// Re-entrancy latch: a kind's New must not trigger another Propose on
	// this DAO (e.g. via a captured handle), which could nest factory
	// calls or grow active storage unboundedly before the first returns.
	if dao.proposing {
		panic("commondao: re-entrant Propose is not allowed")
	}
	dao.proposing = true
	d, err := v.(ProposalKind).New(dao.Readonly(), args)
	dao.proposing = false
	if err != nil {
		return nil, err
	}

	if d == nil {
		return nil, ErrProposalDefinitionRequired
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
// Proposals are re-evaluated after every recorded vote: a YES tally at
// the definition's threshold decides the proposal immediately, and a
// simple majority of NO dismisses it immediately.
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

	if c != ChoiceYes && c != ChoiceNo && c != ChoiceAbstain {
		return ErrInvalidVoteChoice
	}

	p.record.AddVote(Vote{
		addr:   member,
		choice: c,
		reason: reason,
	})

	// Early termination: proposals are decided the moment the outcome is
	// mathematically settled. A passed proposal stays in the active
	// storage until executed; a dismissed one is finished.
	switch TallyDefault(p.record.Readonly(), p.Electorate(), p.definition.Threshold()) {
	case OutcomePassed:
		p.status = StatusPassed
	case OutcomeDismissed:
		dao.dismiss(p)
	}
	return nil
}

// Execute executes a proposal.
//
// Proposals that already passed (decided early by the default Council
// rules) execute immediately. Active proposals are tallied once their
// voting deadline passes and are dismissed unless passed.
//
// sub is the DAO-scoped sub-identity that the host mints and passes into
// the executor as its value-movement authority (see ExecFunc). The
// executor is non-crossing, so it is called directly. Execute itself is
// not a crossing function (sub sits in a non-first parameter slot)
// because /p/ production code cannot declare crossing functions.
func (dao *CommonDAO) Execute(proposalID uint64, sub realm) error {
	if dao.deleted {
		return ErrDAOIsDeleted
	}

	// Re-entrancy latch: an executor must not re-enter Execute on this
	// DAO. Remove-before-run already stops the same proposal from running
	// twice; this additionally blocks an executor from executing a
	// different proposal of the same DAO mid-execution.
	if dao.executing {
		panic("commondao: re-entrant Execute is not allowed")
	}
	dao.executing = true
	defer func() { dao.executing = false }()

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
	// (Validate, the executor) runs, so a re-entrant Execute call
	// cannot run it twice.
	dao.activeProposals.Remove(p.id)

	// Tally proposals that are still active after their deadline;
	// undecided proposals are dismissed. Vote already decides settled
	// outcomes, so this re-tally only matters for definitions whose
	// Threshold is not constant.
	if p.status == StatusActive {
		if TallyDefault(p.record.Readonly(), p.Electorate(), p.definition.Threshold()) == OutcomePassed {
			p.status = StatusPassed
		} else {
			p.status = StatusDismissed
			dao.finishedProposals.Add(p)
			return nil
		}
	}

	// IMPORTANT, from this point on, any error is going to result
	// in a proposal failure and execute will succeed.

	// Validate the passed proposal before execution
	err := p.Validate()

	// Execute proposal only if it's executable
	if err == nil {
		if e, ok := p.Definition().(Executable); ok {
			if fn := e.Executor(); fn != nil {
				err = fn(0, sub)
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
