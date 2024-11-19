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
func WithMember(addr std.Address, roles ...Role) Option {
	return func(dao *AdminDAO) {
		dao.members.Set(addr.String(), append([]Role(nil), roles...))

		for _, r := range roles {
			var (
				members []std.Address
				key     = string(r)
			)

			v, found := dao.roles.Get(key)
			if found {
				members = v.([]std.Address)
			}

			members = append(members, addr)
			dao.roles.Set(key, members)
		}
	}
}
