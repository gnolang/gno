package commondao

import "std"

// Option configures the CommonDAO.
type Option func(*CommonDAO)

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
