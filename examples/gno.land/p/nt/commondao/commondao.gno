package commondao

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/seqid"
)

var (
	ErrInvalidVoteChoice    = errors.New("invalid vote choice")
	ErrMemberExists         = errors.New("member already exist")
	ErrNotMember            = errors.New("account is not a member of the DAO")
	ErrOverflow             = errors.New("next ID overflows uint64")
	ErrProposalNotFound     = errors.New("proposal not found")
	ErrStatusIsNotActive    = errors.New("proposal status is not active")
	ErrVotingDeadlineNotMet = errors.New("voting deadline not met")
)

type (
	// CommonDAO defines a DAO.
	CommonDAO struct {
		id       uint64
		slug     string
		name     string
		parent   *CommonDAO
		members  *avl.Tree // string(std.Address) -> struct{}
		genID    seqid.ID
		active   *avl.Tree // string(proposal ID) -> *Proposal
		finished *avl.Tree // string(proposal ID) -> *Proposal
	}

	// Stats contains proposal voting stats.
	Stats struct {
		YayVotes  int
		NayVotes  int
		Abstained int
	}
)

// New creates a new common DAO.
func New(options ...Option) *CommonDAO {
	dao := &CommonDAO{
		members:  avl.NewTree(),
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

// Parent returns the parent DAO.
// Null can be returned when DAO has no parent assigned.
func (dao CommonDAO) Parent() *CommonDAO {
	return dao.parent
}

// Members returns the list of DAO members.
func (dao CommonDAO) Members() []std.Address {
	var members []std.Address
	dao.members.Iterate("", "", func(key string, _ interface{}) bool {
		members = append(members, std.Address(key))
		return false
	})
	return members
}

// AddMember adds a new member to the DAO.
func (dao *CommonDAO) AddMember(user std.Address) error {
	if dao.IsMember(user) {
		return ErrMemberExists
	}
	dao.members.Set(user.String(), struct{}{})
	return nil
}

// RemoveMember removes a member from the DAO.
func (dao *CommonDAO) RemoveMember(user std.Address) (removed bool) {
	_, removed = dao.members.Remove(user.String())
	return removed
}

// IsMember checks if a user is a member of the DAO.
func (dao CommonDAO) IsMember(user std.Address) bool {
	return dao.members.Has(user.String())
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

// Vote submits a new vote for a proposal.
func (dao *CommonDAO) Vote(member std.Address, proposalID uint64, c VoteChoice) error {
	if c != ChoiceYes && c != ChoiceNo && c != ChoiceAbstain {
		return ErrInvalidVoteChoice
	}

	if !dao.IsMember(member) {
		return ErrNotMember
	}

	p, found := dao.GetActiveProposal(proposalID)
	if !found {
		return ErrProposalNotFound
	}
	return p.record.AddVote(member, c)
}

func (dao *CommonDAO) Tally(p *Proposal) Stats {
	// Initialize stats considering only yes/no votes
	record := p.VotingRecord()
	stats := Stats{
		YayVotes: record.VoteCount(ChoiceYes),
		NayVotes: record.VoteCount(ChoiceNo),
	}
	votesCount := stats.YayVotes + stats.NayVotes
	membersCount := len(dao.Members())
	stats.Abstained = membersCount - votesCount

	percentage := float64(votesCount) / float64(membersCount)
	if percentage < p.Quorum() {
		p.status = StatusFailed
		p.statusReason = "low participation"
		return stats
	}

	if !p.Definition().Tally(record, membersCount) {
		p.status = StatusFailed
		p.statusReason = "no consensus"
	}

	return stats
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

	// Validate proposal before executing it
	def := p.Definition()
	err := def.Validate()
	if err != nil {
		p.status = StatusFailed
		p.statusReason = err.Error()
	} else {
		// Tally votes and update proposal status
		dao.Tally(p)

		// Execute proposal only when the majority vote wins
		if p.Status() != StatusFailed {
			err = def.Execute()
			if err != nil {
				p.status = StatusFailed
				p.statusReason = err.Error()
			} else {
				p.status = StatusExecuted
			}
		}
	}

	// Whichever the outcome of the validation, tallying
	// and execution consider the proposal finished.
	key := makeProposalKey(p.id)
	dao.active.Remove(key)
	dao.finished.Set(key, p)
	return err
}

func makeProposalKey(id uint64) string {
	return seqid.ID(id).String()
}
