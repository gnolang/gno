package admindao

import (
	"errors"
	"std"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/rotree"
	"gno.land/p/demo/seqid"
)

// DefaultQuorum defines the default quorum required to tally proposal votes.
const DefaultQuorum = 0.51 // 51%

var (
	ErrInvalidAddress             = errors.New("invalid address")
	ErrInvalidVoteChoice          = errors.New("invalid vote choice")
	ErrMemberExists               = errors.New("member already exist")
	ErrNotMember                  = errors.New("account is not a member of the DAO")
	ErrOverflow                   = errors.New("next ID overflows uint64")
	ErrProposalDefinitionRequired = errors.New("proposal definition is required")
	ErrProposalNotFound           = errors.New("proposal not found")
	ErrStatusIsNotActive          = errors.New("proposal status is not active")
	ErrVotingDeadlineNotMet       = errors.New("voting deadline not met")
)

type (
	// AdminDAO defines a Boards administration DAO.
	AdminDAO struct {
		parent   *AdminDAO
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

// New creates a new admin DAO.
func New(options ...Option) *AdminDAO {
	dao := &AdminDAO{
		members:  avl.NewTree(),
		active:   avl.NewTree(),
		finished: avl.NewTree(),
	}
	for _, apply := range options {
		apply(dao)
	}
	return dao
}

// Parent returns the parent DAO.
// Null can be returned when DAO has no parent assigned.
func (dao AdminDAO) Parent() *AdminDAO {
	return dao.parent
}

// Members returns the list of DAO members.
func (dao AdminDAO) Members() []std.Address {
	var members []std.Address
	dao.members.Iterate("", "", func(key string, _ interface{}) bool {
		members = append(members, std.Address(key))
		return false
	})
	return members
}

// AddMember adds a new member to the DAO.
func (dao *AdminDAO) AddMember(user std.Address) error {
	if dao.IsMember(user) {
		return ErrMemberExists
	}
	dao.members.Set(user.String(), struct{}{})
	return nil
}

// RemoveMember removes a member from the DAO.
func (dao *AdminDAO) RemoveMember(user std.Address) (removed bool) {
	_, removed = dao.members.Remove(user.String())
	return removed
}

// IsMember checks if a user is a member of the DAO.
func (dao AdminDAO) IsMember(user std.Address) bool {
	return dao.members.Has(user.String())
}

// ActiveProposals returns all active DAO proposals.
func (dao AdminDAO) ActiveProposals() rotree.IReadOnlyTree {
	return dao.active
}

// FinishedProposalsi returns all finished DAO proposals.
func (dao AdminDAO) FinishedProposals() rotree.IReadOnlyTree {
	return dao.finished
}

// Propose creates a new DAO proposal.
func (dao *AdminDAO) Propose(creator std.Address, d ProposalDefinition) (*Proposal, error) {
	if d == nil {
		return nil, ErrProposalDefinitionRequired
	}

	if !creator.IsValid() {
		return nil, ErrInvalidAddress
	}

	id, ok := dao.genID.TryNext()
	if !ok {
		return nil, ErrOverflow
	}

	return NewProposal(uint64(id), creator, d), nil
}

// GetActiveProposal returns an active proposal.
func (dao AdminDAO) GetActiveProposal(proposalID uint64) (_ *Proposal, found bool) {
	key := makeProposalKey(proposalID)
	if v, ok := dao.active.Get(key); ok {
		return v.(*Proposal), true
	}
	return nil, false
}

// GetFinishedProposal returns a finished proposal.
func (dao AdminDAO) GetFinishedProposal(proposalID uint64) (_ *Proposal, found bool) {
	key := makeProposalKey(proposalID)
	if v, ok := dao.finished.Get(key); ok {
		return v.(*Proposal), true
	}
	return nil, false
}

// Vote submits a new vote for a proposal.
func (dao *AdminDAO) Vote(member std.Address, proposalID uint64, c VoteChoice) error {
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

func (dao *AdminDAO) Tally(p *Proposal) Stats {
	// Initialize stats considering only yes/no votes
	record := p.VotingRecord()
	votesCount := record.VoteCount(ChoiceYes) + record.VoteCount(ChoiceNo)
	membersCount := len(dao.Members())
	stats := Stats{
		YayVotes:  record.VoteCount(ChoiceYes),
		NayVotes:  record.VoteCount(ChoiceNo),
		Abstained: membersCount - votesCount,
	}

	// Check quorum before tallying
	quorum := p.Definition().Quorum()
	if quorum <= 0 || quorum > 1 {
		quorum = DefaultQuorum
	}

	percentage := float64(votesCount) / float64(membersCount)
	if percentage < quorum {
		p.status = StatusFailed
		p.statusReason = "low participation"
		return stats
	}

	// Tally by majority requiring more than 50% of yes/no votes to win
	// TODO: Confirm this way of tallying is enough for now. Keeping it minimal.
	choice := record.GetProvableMajorityChoice()
	if record.VoteCount(choice) <= int(votesCount/2) {
		p.status = StatusFailed
		p.statusReason = "majority not met"
	}
	return stats
}

// Execute executes a proposal.
func (dao *AdminDAO) Execute(proposalID uint64) error {
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
