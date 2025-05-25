package commondao

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl/list"
	"gno.land/p/demo/seqid"
)

// PathSeparator is the separator character used in DAO paths.
const PathSeparator = "/"

var (
	ErrInvalidVoteChoice    = errors.New("invalid vote choice")
	ErrNotMember            = errors.New("account is not a member of the DAO")
	ErrOverflow             = errors.New("next ID overflows uint64")
	ErrProposalFailed       = errors.New("proposal failed to pass")
	ErrProposalNotFound     = errors.New("proposal not found")
	ErrVotingDeadlineNotMet = errors.New("voting deadline not met")
)

// CommonDAO defines a DAO.
type CommonDAO struct {
	id                uint64
	slug              string
	name              string
	description       string
	parent            *CommonDAO
	children          list.IList
	members           MemberStorage
	genID             seqid.ID
	activeProposals   ProposalStorage
	finishedProposals ProposalStorage
	deleted           bool // Soft delete
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
func (dao *CommonDAO) Propose(creator std.Address, d ProposalDefinition) (*Proposal, error) {
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
func (dao *CommonDAO) MustPropose(creator std.Address, d ProposalDefinition) *Proposal {
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

// Vote submits a new vote for a proposal.
func (dao *CommonDAO) Vote(member std.Address, proposalID uint64, c VoteChoice, reason string) error {
	if !dao.Members().Has(member) {
		return ErrNotMember
	}

	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
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

// Tally counts votes and validates if a proposal passes.
func (dao *CommonDAO) Tally(proposalID uint64) (passes bool, _ error) {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return false, ErrProposalNotFound
	}

	if p.Status() != StatusActive {
		return false, ErrStatusIsNotActive
	}

	if err := dao.checkProposalPasses(p); err != nil {
		// Don't return an error if proposal failed to pass when tallying
		if err == ErrProposalFailed {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Execute executes a proposal.
func (dao *CommonDAO) Execute(proposalID uint64) error {
	p := dao.activeProposals.Get(proposalID)
	if p == nil {
		return ErrProposalNotFound
	}

	if p.Status() != StatusActive {
		return ErrStatusIsNotActive
	}

	if time.Now().Before(p.VotingDeadline()) {
		return ErrVotingDeadlineNotMet
	}

	// From this point any error results in a proposal failure and successful execution
	err := p.Validate()

	if err == nil {
		err = dao.checkProposalPasses(p)
	}

	if err == nil {
		// Execute proposal only if it's executable
		if e, ok := p.Definition().(Executable); ok {
			err = cross(e.Execute)()
		}
	}

	// Proposal fails if there is any error during validation and execution process
	if err != nil {
		p.status = StatusFailed
		p.statusReason = err.Error()
	} else {
		p.status = StatusPassed
	}

	// Whichever the outcome of the validation, tallying
	// and execution consider the proposal finished.
	dao.activeProposals.Remove(p.id)
	dao.finishedProposals.Add(p)
	return nil
}

func (dao *CommonDAO) checkProposalPasses(p *Proposal) error {
	record := p.VotingRecord().Readonly()
	members := NewMemberSet(dao.Members())
	passes, err := p.Definition().Tally(record, members)
	if err != nil {
		return err
	}

	if !passes {
		return ErrProposalFailed
	}
	return nil
}
