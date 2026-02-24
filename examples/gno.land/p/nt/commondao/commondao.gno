package commondao

import (
	"errors"

	"gno.land/p/nt/avl/list"
	"gno.land/p/nt/seqid"
)

// PathSeparator is the separator character used in DAO paths.
const PathSeparator = "/"

var (
	ErrExecutionNotAllowed  = errors.New("proposal must pass before execution")
	ErrInvalidVoteChoice    = errors.New("invalid vote choice")
	ErrNotMember            = errors.New("account is not a member of the DAO")
	ErrOverflow             = errors.New("next ID overflows uint64")
	ErrProposalNotFound     = errors.New("proposal not found")
	ErrVotingDeadlineNotMet = errors.New("voting deadline not met")
	ErrVotingDeadlinePassed = errors.New("voting deadline has passed")
	ErrWithdrawalNotAllowed = errors.New("withdrawal not allowed for proposals with votes")
)

// CommonDAO defines a DAO.
type CommonDAO struct {
	id                         uint64
	slug                       string
	name                       string
	description                string
	parent                     *CommonDAO
	children                   list.IList
	members                    MemberStorage
	genID                      seqid.ID
	activeProposals            ProposalStorage
	finishedProposals          ProposalStorage
	deleted                    bool // Soft delete
	disableVotingDeadlineCheck bool
}

// New creates a new common DAO.
func New(options ...Option) *CommonDAO {
	dao := &CommonDAO{
		children:          &list.List{},
		members:           NewMemberStorage(),
		activeProposals:   NewProposalStorage(),
		finishedProposals: NewProposalStorage(),
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

// Slug returns DAO's URL slug.
func (dao CommonDAO) Slug() string {
	return dao.slug
}

// Name returns DAO's name.
func (dao CommonDAO) Name() string {
	return dao.name
}

// Description returns DAO's description.
func (dao CommonDAO) Description() string {
	return dao.description
}

// Path returns the full path to the DAO.
// Paths are normally used when working with hierarchical
// DAOs and is created by concatenating DAO slugs.
func (dao CommonDAO) Path() string {
	// NOTE: Path could be a value but there might be use cases where dynamic path is useful (?)
	parent := dao.Parent()
	if parent != nil {
		prefix := parent.Path()
		if prefix != "" {
			return prefix + PathSeparator + dao.slug
		}
	}
	return dao.slug
}

// Parent returns the parent DAO.
// Null can be returned when DAO has no parent assigned.
func (dao CommonDAO) Parent() *CommonDAO {
	return dao.parent
}

// Children returns a list with the direct DAO children.
// Each item in the list is a reference to a CommonDAO instance.
func (dao CommonDAO) Children() list.IList {
	return dao.children
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

// Members returns the list of DAO members.
func (dao CommonDAO) Members() MemberStorage {
	return dao.members
}

// ActiveProposals returns active DAO proposals.
func (dao CommonDAO) ActiveProposals() ProposalStorage {
	return dao.activeProposals
}

// FinishedProposalsi returns finished DAO proposals.
func (dao CommonDAO) FinishedProposals() ProposalStorage {
	return dao.finishedProposals
}

// IsDeleted returns true when DAO has been soft deleted.
func (dao CommonDAO) IsDeleted() bool {
	return dao.deleted
}

// SetDeleted changes DAO's soft delete flag.
func (dao *CommonDAO) SetDeleted(deleted bool) {
	dao.deleted = deleted
}

// Propose creates a new DAO proposal.
func (dao *CommonDAO) Propose(creator address, d ProposalDefinition) (*Proposal, error) {
	id, ok := dao.genID.TryNext()
	if !ok {
		return nil, ErrOverflow
	}

	p, err := NewProposal(uint64(id), creator, d)
	if err != nil {
		return nil, err
	}

	dao.activeProposals.Add(p)
	return p, nil
}

// MustPropose creates a new DAO proposal or panics on error.
func (dao *CommonDAO) MustPropose(creator address, d ProposalDefinition) *Proposal {
	p, err := dao.Propose(creator, d)
	if err != nil {
		panic(err)
	}
	return p
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
// Only proposals without votes can be withdrawn, and once
// withdrawn they are considered finished.
func (dao *CommonDAO) Withdraw(proposalID uint64) error {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	if p.VotingRecord().Size() > 0 {
		return ErrWithdrawalNotAllowed
	}

	p.status = StatusWithdrawn
	dao.activeProposals.Remove(p.id)
	dao.finishedProposals.Add(p)
	return nil
}

// Vote submits a new vote for a proposal.
//
// By default votes are only allowed to members of the DAO when the proposal is active,
// and within the voting period. No votes are allowed once the voting deadline passes.
// DAO deadline checks can optionally be disabled using the `DisableVotingDeadlineCheck` option.
func (dao *CommonDAO) Vote(member address, proposalID uint64, c VoteChoice, reason string) error {
	if !dao.Members().Has(member) {
		return ErrNotMember
	}

	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	if !dao.disableVotingDeadlineCheck && p.HasVotingDeadlinePassed() {
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
	return nil
}

// Execute executes a proposal.
//
// By default active proposals can only be executed after their voting deadline passes.
// DAO deadline checks can optionally be disabled using the `DisableVotingDeadlineCheck` option.
func (dao *CommonDAO) Execute(proposalID uint64) error {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	// Proposal must be active or have passed to be executed
	if p.status != StatusActive && p.status != StatusPassed {
		return ErrExecutionNotAllowed
	}

	// Execution must be done after voting deadline
	if !dao.disableVotingDeadlineCheck && !p.HasVotingDeadlinePassed() {
		return ErrVotingDeadlineNotMet
	}

	// IMPORTANT, from this point on, any error is going to result
	// in a proposal failure and execute will succeed.

	// Validate proposal before execution
	err := p.Validate()

	// Tally votes and update proposal status to "passed" or "rejected"
	if err == nil {
		err = p.Tally(dao.Members())
		if err == nil && p.Status() == StatusRejected {
			// Don't try to execute proposal if it's been rejected
			return nil
		}
	}

	// Execute proposal only if it's executable
	if err == nil {
		if e, ok := p.Definition().(Executable); ok {
			if fn := e.Executor(); fn != nil {
				err = fn(cross)
			}
		}
	}

	// Proposal fails if there is any error during validation and execution process
	if err != nil {
		p.status = StatusFailed
		p.statusReason = err.Error()
	} else {
		p.status = StatusExecuted
	}

	// Whichever the outcome of the validation, tallying
	// and execution consider the proposal finished.
	dao.activeProposals.Remove(p.id)
	dao.finishedProposals.Add(p)
	return nil
}
