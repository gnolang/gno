package commondao

import "std"

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
func WithMember(addr std.Address) Option {
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
