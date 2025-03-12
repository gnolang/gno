package commondao

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/addrset"
)

// PathSeparator is the separator character used in DAO paths.
const PathSeparator = "/"

var (
	ErrInvalidVoteChoice    = errors.New("invalid vote choice")
	ErrLowParticipation     = errors.New("low participation")
	ErrNoConcensus          = errors.New("no concensus")
	ErrNotMember            = errors.New("account is not a member of the DAO")
	ErrOverflow             = errors.New("next ID overflows uint64")
	ErrProposalNotFound     = errors.New("proposal not found")
	ErrVotingDeadlineNotMet = errors.New("voting deadline not met")
)

// CommonDAO defines a DAO.
type CommonDAO struct {
	id       uint64
	slug     string
	name     string
	parent   *CommonDAO
	members  *addrset.Set
	genID    seqid.ID
	active   *avl.Tree // string(proposal ID) -> *Proposal
	finished *avl.Tree // string(proposal ID) -> *Proposal
}

// New creates a new common DAO.
func New(options ...Option) *CommonDAO {
	dao := &CommonDAO{
		members:  &addrset.Set{},
		active:   avl.NewTree(),
		finished: avl.NewTree(),
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

// Path returns the full path to the DAO.
// Paths are normally used when working with hierarchical
// DAOs and is created by concatenating DAO slugs.
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
func (dao CommonDAO) Members() *addrset.Set {
	return dao.members
}

// ActiveProposals returns all active DAO proposals.
func (dao CommonDAO) ActiveProposals() rotree.IReadOnlyTree {
	return dao.active
}

// FinishedProposalsi returns all finished DAO proposals.
func (dao CommonDAO) FinishedProposals() rotree.IReadOnlyTree {
	return dao.finished
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

	key := makeProposalKey(p.ID())
	dao.active.Set(key, p)
	return p, nil
}

// GetActiveProposal returns an active proposal.
func (dao CommonDAO) GetActiveProposal(proposalID uint64) (_ *Proposal, found bool) {
	key := makeProposalKey(proposalID)
	if v, ok := dao.active.Get(key); ok {
		return v.(*Proposal), true
	}
	return nil, false
}

// GetFinishedProposal returns a finished proposal.
func (dao CommonDAO) GetFinishedProposal(proposalID uint64) (_ *Proposal, found bool) {
	key := makeProposalKey(proposalID)
	if v, ok := dao.finished.Get(key); ok {
		return v.(*Proposal), true
	}
	return nil, false
}

// GetProposal returns an proposal.
func (dao CommonDAO) GetProposal(proposalID uint64) (p *Proposal, found bool) {
	p, found = dao.GetActiveProposal(proposalID)
	if !found {
		return dao.GetFinishedProposal(proposalID)
	}
	return p, found
}

// Vote submits a new vote for a proposal.
func (dao *CommonDAO) Vote(member std.Address, proposalID uint64, c VoteChoice, reason string) error {
	// TODO: Support other voting options though the proposal definition
	if c != ChoiceYes && c != ChoiceNo && c != ChoiceAbstain {
		return ErrInvalidVoteChoice
	}

	if !dao.Members().Has(member) {
		return ErrNotMember
	}

	p, found := dao.GetActiveProposal(proposalID)
	if !found {
		return ErrProposalNotFound
	}

	p.record.AddVote(Vote{
		Address: member,
		Choice:  c,
		Reason:  reason,
	})
	return nil
}

func (dao *CommonDAO) Tally(proposalID uint64) error { // TODO: Rethink or refactor tally to return the winning choice
	p, found := dao.GetActiveProposal(proposalID)
	if !found {
		return ErrProposalNotFound
	}

	if p.Status() != StatusActive {
		return ErrStatusIsNotActive
	}

	return dao.tallyProposal(p)
}

// Execute executes a proposal.
func (dao *CommonDAO) Execute(proposalID uint64) error {
	p, found := dao.GetActiveProposal(proposalID)
	if !found {
		return ErrProposalNotFound
	}

	if p.Status() != StatusActive {
		return ErrStatusIsNotActive
	}

	if time.Now().Before(p.VotingDeadline()) {
		return ErrVotingDeadlineNotMet
	}

	// From this point any error results in execution success and proposal failure
	err := p.Validate()
	if err == nil {
		err = dao.executeProposal(p)
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
	key := makeProposalKey(p.id)
	dao.active.Remove(key)
	dao.finished.Set(key, p)
	return nil
}

func (dao *CommonDAO) tallyProposal(p *Proposal) error {
	// TODO: Support other voting options though the proposal definition
	record := p.VotingRecord()
	votesCount := record.VoteCount(ChoiceYes) + record.VoteCount(ChoiceNo) // Only consider YES/NO votes
	membersCount := dao.Members().Size()
	percentage := float64(votesCount) / float64(membersCount)
	if percentage < p.Quorum() {
		return ErrLowParticipation
	}

	if !p.Definition().Tally(record.ReadOnly(), membersCount) {
		return ErrNoConcensus
	}
	return nil
}

func (dao *CommonDAO) executeProposal(p *Proposal) error {
	if p.Status() != StatusActive {
		return ErrStatusIsNotActive
	}

	if err := dao.tallyProposal(p); err != nil {
		return err
	}

	if e, ok := p.Definition().(Executable); ok {
		return e.Execute()
	}
	return nil
}

func makeProposalKey(id uint64) string {
	return seqid.ID(id).String()
}
