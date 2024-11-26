package boards2

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/boards2/admindao"
)

// TODO: Support to deal with permissions for anonymous users?

const (
	RoleOwner     Role = "owner"
	RoleAdmin          = "admin"
	RoleModerator      = "moderator"
)

type (
	// Role defines the type for user roles.
	Role string

	// DefaultPermissions manages users, roles and permissions.
	DefaultPermissions struct {
		superRole Role
		dao       *admindao.AdminDAO
		users     *avl.Tree // string(std.Address) -> []Role
		roles     *avl.Tree // string(role) -> []Permission
	}
)

// NewDefaultPermissions creates a new permissions type.
// This type is a default implementation to handle users, roles and permissions.
func NewDefaultPermissions(dao *admindao.AdminDAO, options ...DefaultPermissionsOption) *DefaultPermissions {
	dp := &DefaultPermissions{
		dao:   dao,
		roles: avl.NewTree(),
		users: avl.NewTree(),
	}
	for _, apply := range options {
		apply(dp)
	}
	return dp
}

// Roles returns the list of roles.
func (dp DefaultPermissions) Roles() []Role {
	var roles []Role
	dp.roles.Iterate("", "", func(name string, _ interface{}) bool {
		roles = append(roles, Role(name))
		return false
	})
	return roles
}

// GetUserRoles returns the list of roles assigned to a user.
func (dp DefaultPermissions) GetUserRoles(user std.Address) []Role {
	v, found := dp.users.Get(user.String())
	if !found {
		return nil
	}
	return v.([]Role)
}

// HasRole checks if a user has a specific role assigned.
func (dp DefaultPermissions) HasRole(user std.Address, r Role) bool {
	for _, role := range dp.GetUserRoles(user) {
		if role == r {
			return true
		}
	}
	return false
}

// HasPermission checks if a user has a specific permission.
func (dp DefaultPermissions) HasPermission(user std.Address, perm Permission) bool {
	// TODO: Should we check that the user belongs to the DAO?
	for _, r := range dp.GetUserRoles(user) {
		v, found := dp.roles.Get(string(r))
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
func (dp DefaultPermissions) WithPermission(user std.Address, perm Permission, a Args, cb func(Args)) {
	if !dp.HasPermission(user, perm) || !dp.dao.IsMember(user) {
		panic("unauthorized")
	}

	// TODO: Support DAO proposals that run the callback on proposal execution
	cb(a)
}
