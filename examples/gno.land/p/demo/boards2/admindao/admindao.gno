package admindao

import (
	"std"

	"gno.land/p/demo/avl"
)

// TODO: Add support for proposals
// TODO: Add support for events

type (
	// Role defines a DAO user role.
	Role string

	// Member defines a DAO member.
	Member struct {
		Address std.Address
		Roles   []Role
	}

	// AdminDAO defines a Boards administration DAO.
	AdminDAO struct {
		parent  *AdminDAO
		members *avl.Tree // string(std.Address) -> []Role
		roles   *avl.Tree // string(role) -> []std.Address
	}
)

// New creates a new admin DAO.
func New(options ...Option) *AdminDAO {
	dao := &AdminDAO{
		members: avl.NewTree(),
		roles:   avl.NewTree(),
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
func (dao AdminDAO) Members() []Member {
	var members []Member
	dao.members.Iterate("", "", func(key string, v interface{}) bool {
		members = append(members, Member{
			Address: std.Address(key),
			Roles:   v.([]Role),
		})
		return false
	})
	return members
}

// Roles returns the list of DAO roles.
func (dao AdminDAO) Roles() []Role {
	var roles []Role
	dao.roles.Iterate("", "", func(name string, _ interface{}) bool {
		roles = append(roles, Role(name))
		return false
	})
	return roles
}

// GetMemberRoles returns the list of roles assigned to a DAO member.
func (dao AdminDAO) GetMemberRoles(member std.Address) []Role {
	v, found := dao.members.Get(member.String())
	if !found {
		return nil
	}
	return v.([]Role)
}

// IsMember checks if a user is a member of the DAO.
func (dao AdminDAO) IsMember(user std.Address) bool {
	return dao.members.Has(user.String())
}

// HasRole checks if a member has a specific role assigned.
func (dao AdminDAO) HasRole(member std.Address, r Role) bool {
	v, found := dao.members.Get(member.String())
	if !found {
		return false
	}

	for _, role := range v.([]Role) {
		if role == r {
			return true
		}
	}
	return false
}
