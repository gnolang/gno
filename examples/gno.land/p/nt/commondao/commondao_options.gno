package commondao

// Option configures the CommonDAO.
type Option func(*CommonDAO)

// WithID assigns a unique identifier to the DAO.
func WithID(id uint64) Option {
	return func(dao *CommonDAO) {
		dao.id = id
	}
}

// WithName assigns a name to the DAO.
func WithName(name string) Option {
	return func(dao *CommonDAO) {
		dao.name = name
	}
}

// WithDescription assigns a description to the DAO.
func WithDescription(description string) Option {
	return func(dao *CommonDAO) {
		dao.description = description
	}
}

// WithSlug assigns a URL slug to the DAO.
func WithSlug(slug string) Option {
	return func(dao *CommonDAO) {
		dao.slug = slug
	}
}

// WithParent assigns a parent DAO.
func WithParent(p *CommonDAO) Option {
	return func(dao *CommonDAO) {
		dao.parent = p
	}
}

// WithChildren assigns one or more direct child SubDAOs to the DAO.
func WithChildren(children ...*CommonDAO) Option {
	return func(dao *CommonDAO) {
		for _, subDAO := range children {
			dao.children.Append(subDAO)
		}
	}
}

// WithMember assigns a member to the DAO.
func WithMember(addr address) Option {
	return func(dao *CommonDAO) {
		dao.members.Add(addr)
	}
}

// WithMemberStorage assigns a custom member storage to the DAO.
// An empty member storage is used by default if the specified storage is nil.
func WithMemberStorage(s MemberStorage) Option {
	return func(dao *CommonDAO) {
		if s == nil {
			s = NewMemberStorage()
		}
		dao.members = s
	}
}

// WithActiveProposalStorage assigns a custom storage for active proposals.
// A default empty proposal storage is used when the custopm storage is nil.
// Custom storage implementations can be used to store proposals in a different location.
func WithActiveProposalStorage(s ProposalStorage) Option {
	return func(dao *CommonDAO) {
		if s == nil {
			s = NewProposalStorage()
		}
		dao.activeProposals = s
	}
}

// WithFinishedProposalStorage assigns a custom storage for finished proposals.
// A default empty proposal storage is used when the custopm storage is nil.
// Custom storage implementations can be used to store proposals in a different location.
func WithFinishedProposalStorage(s ProposalStorage) Option {
	return func(dao *CommonDAO) {
		if s == nil {
			s = NewProposalStorage()
		}
		dao.finishedProposals = s
	}
}

// DisableVotingDeadlineCheck disables voting deadline check when voting or executing proposals.
// By default CommonDAO checks that the proposal voting deadline has not been met when a new vote
// is submitted, before registering the vote, and on proposal execution it also checks that voting
// deadline has been met before executing a proposal.
//
// Disabling these checks can be useful in different use cases, moving the responsibility to check
// the deadline to the commondao package caller. One example where this could be useful would be
// in case where 100% or a required number of members of a DAO vote on a proposal and reach consensus
// before the deadline is met, otherwise proposal would have to wait until deadline to be executed.
func DisableVotingDeadlineCheck() Option {
	return func(dao *CommonDAO) {
		dao.disableVotingDeadlineCheck = true
	}
}
