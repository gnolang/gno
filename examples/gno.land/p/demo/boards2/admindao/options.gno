package admindao

import "std"

// Option configures the AdminDAO.
type Option func(*AdminDAO)

// WithParent assigns a parent DAO.
func WithParent(p *AdminDAO) Option {
	return func(dao *AdminDAO) {
		dao.parent = p
	}
}

// WithMember assigns a member to the DAO.
func WithMember(addr std.Address) Option {
	return func(dao *AdminDAO) {
		dao.members.Set(addr.String(), struct{}{})
	}
}
