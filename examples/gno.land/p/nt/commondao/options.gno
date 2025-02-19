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

// WithMember assigns a member to the DAO.
func WithMember(addr std.Address) Option {
	return func(dao *CommonDAO) {
		dao.members.Set(addr.String(), struct{}{})
	}
}
