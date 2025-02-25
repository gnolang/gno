package commondao

import (
	"std"

	"gno.land/p/moul/addrset"
)

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
		dao.members.Add(addr)
	}
}

// WithMembers assigns multiple members to the DAO.
// An empty member set is used by default when specified members set is nil.
func WithMembers(members *addrset.Set) Option {
	return func(dao *CommonDAO) {
		if members == nil {
			members = &addrset.Set{}
		}
		dao.members = members
	}
}
