package admindao

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

// TODO: Add support for events

var (
	ErrInvalidAddress             = errors.New("invalid address")
	ErrInvalidVoteChoice          = errors.New("invalid vote choice")
	ErrMemberExists               = errors.New("member already exist")
	ErrNotMember                  = errors.New("account is not a member of the DAO")
	ErrOverflow                   = errors.New("next ID overflows uint64")
	ErrProposalDefinitionRequired = errors.New("proposal definition is required")
	ErrProposalNotFound           = errors.New("proposal not found")
	ErrStatusIsNotActive          = errors.New("proposal status is not active")
	ErrVoteExists                 = errors.New("member already voted")
)

// AdminDAO defines a Boards administration DAO.
type AdminDAO struct {
	parent   *AdminDAO
	members  *avl.Tree // string(std.Address) -> struct{}
	genID    seqid.ID
	active   *avl.Tree // string(proposal ID) -> *Proposal
	finished *avl.Tree // string(proposal ID) -> *Proposal
}

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
func (dao AdminDAO) ActiveProposals() IReadOnlyTree {
	return dao.active
}

// FinishedProposalsi returns all finished DAO proposals.
func (dao AdminDAO) FinishedProposals() IReadOnlyTree {
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

	return &Proposal{
		id:         id,
		status:     StatusActive,
		definition: d,
		creator:    creator,
		votes:      avl.NewTree(),
	}, nil
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

// TODO: Who can vote? All DAO members? Boards should have users and DAO members? (there are no roles here)
// TODO: Do we need to introduce member weight to calculate votes?
// TODO: Default to use majority 20%?

// Vote submits a new vote for a proposal.
func (dao *AdminDAO) Vote(member std.Address, proposalID uint64, c VoteChoice) error {
	if c != ChoiceYes && c != ChoiceNo {
		return ErrInvalidVoteChoice
	}

	if !dao.IsMember(member) {
		return ErrNotMember
	}

	p, found := dao.GetActiveProposal(proposalID)
	if !found {
		return ErrProposalNotFound
	}

	key := member.String()
	if p.votes.Has(key) {
		return ErrVoteExists
	}

	p.votes.Set(key, c)
	return nil
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

	// TODO: check quorum and tally

	def := p.Definition()
	err := def.Validate()
	if err == nil {
		err = def.Execute()
	}

	if err != nil {
		p.status = StatusFailed
		p.failReason = err.Error()
	} else {
		p.status = StatusExecuted
	}

	key := p.id.String()
	dao.active.Remove(key)
	dao.finished.Set(key, p)
	return err
}

func makeProposalKey(id uint64) string {
	return seqid.ID(id).String()
}
