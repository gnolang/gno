package boards2

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/boards2/admindao"
)

const (
	RoleOwner     Role = "owner"
	RoleAdmin          = "admin"
	RoleModerator      = "moderator"
)

type (
	// Role defines the type for user roles.
	Role string

	// ACL or access control list manages user roles and permissions.
	ACL struct {
		superRole Role
		dao       *admindao.AdminDAO
		users     *avl.Tree // string(std.Address) -> []Role
		roles     *avl.Tree // string(role) -> []Permission
	}
)

// NewACL create a new access control list.
func NewACL(dao *admindao.AdminDAO, options ...ACLOption) *ACL {
	acl := &ACL{
		dao:   dao,
		roles: avl.NewTree(),
		users: avl.NewTree(),
	}
	for _, apply := range options {
		apply(acl)
	}
	return acl
}

// Roles returns the list of roles.
func (acl ACL) Roles() []Role {
	var roles []Role
	acl.roles.Iterate("", "", func(name string, _ interface{}) bool {
		roles = append(roles, Role(name))
		return false
	})
	return roles
}

// GetUserRoles returns the list of roles assigned to a user.
func (acl ACL) GetUserRoles(user std.Address) []Role {
	v, found := acl.users.Get(user.String())
	if !found {
		return nil
	}
	return v.([]Role)
}

// HasRole checks if a user has a specific role assigned.
func (acl ACL) HasRole(user std.Address, r Role) bool {
	for _, role := range acl.GetUserRoles(user) {
		if role == r {
			return true
		}
	}
	return false
}

// HasPermission checks if a user has a specific permission.
func (acl ACL) HasPermission(user std.Address, perm Permission) bool {
	for _, r := range acl.GetUserRoles(user) {
		v, found := acl.roles.Get(string(r))
		if !found {
			continue
		}

		for _, p := range v.([]Permission) {
			if p == perm {
				return true
			}
		}
	}
	return false
}

// WithPermission calls a callback when a user has a specific permission.
// It panics on error.
func (acl ACL) WithPermission(user std.Address, perm Permission, a Args, cb func(Args)) {
	if !acl.HasPermission(user, perm) || !acl.dao.IsMember(user) {
		panic("unauthorized")
	}

	// TODO: Support DAO proposals that run the callback on proposal execution

	cb(a)
}
