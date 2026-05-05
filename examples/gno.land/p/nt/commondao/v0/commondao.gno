package commondao

import (
	"errors"

	"gno.land/p/nt/bptree/v0/list"
	"gno.land/p/nt/seqid/v0"
)

// PathSeparator is the separator character used in DAO paths.
const PathSeparator = "/"

var (
	ErrExecutionNotAllowed  = errors.New("execution not allowed, proposal status must be active, passed or rejected")
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
// Paths are normally used when working with hierarchical DAOs and is created
// by concatenating DAO slugs. Method doesn't validate that DAOs have a slug,
// is up to the caller to make sure all DAOs within a path have a valid slug.
func (dao CommonDAO) Path() string {
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

// FinishedProposals returns finished DAO proposals.
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
// Only proposals without votes can be withdrawn, and once withdrawn they are considered finished.
func (dao *CommonDAO) Withdraw(proposalID uint64) error {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	if p.VotingRecord().Size() > 0 {
		return ErrWithdrawalNotAllowed
	}

	p.status = StatusWithdrawn
	dao.finishProposal(p)
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

// Tally counts votes and updates proposal status with the current outcome.
// Proposal status is updated to "passed" when proposal is approved
// or to "rejected" if proposal doesn't pass.
//
// It optionally can validate the proposal to verify potential issues with the
// current state when the proposal type supports validation.
func (dao *CommonDAO) Tally(proposalID uint64, validate bool) (ProposalStatus, error) {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return "", ErrProposalNotFound
	}

	if p.status != StatusActive && p.status != StatusPassed && p.status != StatusRejected {
		return "", ErrTallyNotAllowed
	}

	if validate {
		if err := p.Validate(); err != nil {
			return "", err
		}
	}

	if err := p.tally(dao.members); err != nil {
		return "", err
	}
	return p.status, nil
}

// Execute executes a proposal.
//
// It first validates the proposal, then tallies the votes and finally executes the proposal.
// Proposal is stored as finished after execution, independently of the final status.
//
// By default proposals can only be executed after their voting deadline passes.
// DAO deadline checks can optionally be disabled using the `DisableVotingDeadlineCheck` option.
func (dao *CommonDAO) Execute(proposalID uint64) error {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	// Check that current proposal status is valid for execution. Proposals can be tallied
	// before execution, so its status can change before execution, depending on how the
	// commondao package caller implements the DAO realm.
	if p.status != StatusActive && p.status != StatusPassed && p.status != StatusRejected {
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
		err = p.tally(dao.Members())
		if err == nil && p.Status() == StatusRejected {
			// Don't execute proposal if it's been rejected
			dao.finishProposal(p)
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

	// Whichever the outcome consider the proposal finished
	dao.finishProposal(p)
	return nil
}

func (dao *CommonDAO) finishProposal(p *Proposal) {
	dao.activeProposals.Remove(p.id)
	dao.finishedProposals.Add(p)
}
