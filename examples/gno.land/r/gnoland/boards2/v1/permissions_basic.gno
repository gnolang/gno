package boards2

import (
	"std"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/commondao"
)

// ValidatorFunc defines a function type for permissions validators.
type ValidatorFunc func(Permissions, Args) error

// BasicPermissions manages users, roles and permissions.
type BasicPermissions struct {
	superRole  Role
	dao        *commondao.CommonDAO
	users      *avl.Tree // string(std.Address) -> []Role
	roles      *avl.Tree // string(role) -> []Permission
	validators *avl.Tree // string(Permission) -> BasicPermissionValidator
}

// NewBasicPermissions creates a new permissions type.
// This type is a default implementation to handle users, roles and permissions.
func NewBasicPermissions(dao *commondao.CommonDAO) *BasicPermissions {
	if dao == nil {
		panic("basic permissions require a DAO")
	}

	return &BasicPermissions{
		dao:        dao,
		roles:      avl.NewTree(),
		users:      avl.NewTree(),
		validators: avl.NewTree(),
	}
}

// ValidateFunc add a validator function for a permission.
func (bp *BasicPermissions) ValidateFunc(p Permission, fn ValidatorFunc) {
	bp.validators.Set(string(p), fn)
}

// SetSuperRole assigns a super role.
// A super role is one that have all permissions.
// These type of role doesn't need to be mapped to any permission.
func (bp *BasicPermissions) SetSuperRole(r Role) {
	bp.superRole = r
}

// AddRole add a role with one or more assigned permissions.
func (bp *BasicPermissions) AddRole(r Role, p Permission, extra ...Permission) {
	bp.roles.Set(string(r), append([]Permission{p}, extra...))
}

// RoleExists checks if a role exists.
func (bp BasicPermissions) RoleExists(r Role) bool {
	return (bp.superRole != "" && r == bp.superRole) || bp.roles.Has(string(r))
}

// GetUserRoles returns the list of roles assigned to a user.
func (bp BasicPermissions) GetUserRoles(user std.Address) []Role {
	v, found := bp.users.Get(user.String())
	if !found {
		return nil
	}
	return v.([]Role)
}

// HasRole checks if a user has a specific role assigned.
func (bp BasicPermissions) HasRole(user std.Address, r Role) bool {
	for _, role := range bp.GetUserRoles(user) {
		if role == r {
			return true
		}
	}
	return false
}

// HasPermission checks if a user has a specific permission.
func (bp BasicPermissions) HasPermission(user std.Address, perm Permission) bool {
	for _, r := range bp.GetUserRoles(user) {
		if bp.superRole == r {
			return true
		}

		v, found := bp.roles.Get(string(r))
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

// SetUserRoles adds a new user when it doesn't exist and sets its roles.
// Method can also be called to change the roles of an existing user.
// All user's roles can be removed by calling this method without roles.
func (bp *BasicPermissions) SetUserRoles(_ realm, user std.Address, roles ...Role) {
	if !bp.HasUser(user) {
		bp.dao.Members().Add(user)
	}

	for _, r := range roles {
		if !bp.RoleExists(r) {
			panic("invalid role: " + string(r))
		}
	}

	bp.users.Set(user.String(), append([]Role(nil), roles...))
}

// RemoveUser removes a user from permissions.
func (bp *BasicPermissions) RemoveUser(_ realm, user std.Address) bool {
	_, removed := bp.users.Remove(user.String())
	bp.dao.Members().Remove(user)
	return removed
}

// HasUser checks if a user exists.
func (bp BasicPermissions) HasUser(user std.Address) bool {
	return bp.dao.Members().Has(user)
}

// UsersCount returns the total number of users the permissioner contains.
func (bp BasicPermissions) UsersCount() int {
	return bp.dao.Members().Size()
}

// IterateUsers iterates permissions' users.
func (bp BasicPermissions) IterateUsers(start, count int, fn UsersIterFn) (stopped bool) {
	bp.dao.Members().IterateByOffset(start, count, func(addr std.Address) bool {
		roles := bp.GetUserRoles(addr)
		stopped = fn(User{
			Address: addr,
			Roles:   roles,
		})
		return stopped
	})
	return
}

// WithPermission calls a callback when a user has a specific permission.
// It panics on error or when a handler panics.
// Callbacks are by default called when there is no handle registered for the permission.
func (bp *BasicPermissions) WithPermission(_ realm, user std.Address, p Permission, args Args, cb func(realm, Args)) {
	if !bp.HasPermission(user, p) || !bp.dao.Members().Has(user) {
		panic("unauthorized")
	}

	v, found := bp.validators.Get(string(p))
	if found {
		err := v.(ValidatorFunc)(bp, args)
		if err != nil {
			panic(err)
		}
	}

	cb(cross, args)
}

func createBasicPermissions(owners ...std.Address) *BasicPermissions {
	perms := NewBasicPermissions(commondao.New())
	perms.SetSuperRole(RoleOwner)
	perms.AddRole(RoleAdmin, PermissionBoardCreate)
	for _, owner := range owners {
		perms.SetUserRoles(cross, owner, RoleOwner)
	}
	return perms
}
