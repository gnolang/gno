package boards2

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/commondao"
)

// ValidatorFunc defines a function type for permissions validators.
type ValidatorFunc func(boards.Permissions, boards.Args) error

// BasicPermissions manages users, roles and permissions.
type BasicPermissions struct {
	superRole  boards.Role
	dao        *commondao.CommonDAO
	users      *avl.Tree // string(address) -> []boards.Role
	roles      *avl.Tree // string(role) -> []boards.Permission
	validators *avl.Tree // string(boards.Permission) -> BasicPermissionValidator
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
func (bp *BasicPermissions) ValidateFunc(p boards.Permission, fn ValidatorFunc) {
	bp.validators.Set(string(p), fn)
}

// SetSuperRole assigns a super role.
// A super role is one that have all permissions.
// These type of role doesn't need to be mapped to any permission.
func (bp *BasicPermissions) SetSuperRole(r boards.Role) {
	bp.superRole = r
}

// AddRole add a role with one or more assigned permissions.
func (bp *BasicPermissions) AddRole(r boards.Role, p boards.Permission, extra ...boards.Permission) {
	bp.roles.Set(string(r), append([]boards.Permission{p}, extra...))
}

// RoleExists checks if a role exists.
func (bp BasicPermissions) RoleExists(r boards.Role) bool {
	return (bp.superRole != "" && r == bp.superRole) || bp.roles.Has(string(r))
}

// GetUserRoles returns the list of roles assigned to a user.
func (bp BasicPermissions) GetUserRoles(user address) []boards.Role {
	v, found := bp.users.Get(user.String())
	if !found {
		return nil
	}
	return v.([]boards.Role)
}

// HasRole checks if a user has a specific role assigned.
func (bp BasicPermissions) HasRole(user address, r boards.Role) bool {
	for _, role := range bp.GetUserRoles(user) {
		if role == r {
			return true
		}
	}
	return false
}

// HasPermission checks if a user has a specific permission.
func (bp BasicPermissions) HasPermission(user address, perm boards.Permission) bool {
	for _, r := range bp.GetUserRoles(user) {
		if bp.superRole == r {
			return true
		}

		v, found := bp.roles.Get(string(r))
		if !found {
			continue
		}

		for _, p := range v.([]boards.Permission) {
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
func (bp *BasicPermissions) SetUserRoles(_ realm, user address, roles ...boards.Role) {
	if !bp.HasUser(user) {
		bp.dao.Members().Add(user)
	}

	for _, r := range roles {
		if !bp.RoleExists(r) {
			panic("invalid role: " + string(r))
		}
	}

	bp.users.Set(user.String(), append([]boards.Role(nil), roles...))
}

// RemoveUser removes a user from permissions.
func (bp *BasicPermissions) RemoveUser(_ realm, user address) bool {
	_, removed := bp.users.Remove(user.String())
	bp.dao.Members().Remove(user)
	return removed
}

// HasUser checks if a user exists.
func (bp BasicPermissions) HasUser(user address) bool {
	return bp.dao.Members().Has(user)
}

// UsersCount returns the total number of users the permissioner contains.
func (bp BasicPermissions) UsersCount() int {
	return bp.dao.Members().Size()
}

// IterateUsers iterates permissions' users.
func (bp BasicPermissions) IterateUsers(start, count int, fn boards.UsersIterFn) (stopped bool) {
	bp.dao.Members().IterateByOffset(start, count, func(addr address) bool {
		roles := bp.GetUserRoles(addr)
		stopped = fn(boards.User{
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
func (bp *BasicPermissions) WithPermission(_ realm, user address, p boards.Permission, args boards.Args, cb func(realm)) {
	if !bp.HasPermission(user, p) || !bp.dao.Members().Has(user) {
		panic("unauthorized")
	}

	// Execute custom validation before calling the callback
	v, found := bp.validators.Get(string(p))
	if found {
		err := v.(ValidatorFunc)(bp, args)
		if err != nil {
			panic(err)
		}
	}

	cb(cross)
}

func createBasicPermissions(owners ...address) *BasicPermissions {
	perms := NewBasicPermissions(commondao.New())
	perms.SetSuperRole(RoleOwner)
	perms.AddRole(RoleAdmin, PermissionBoardCreate)
	for _, owner := range owners {
		perms.SetUserRoles(cross, owner, RoleOwner)
	}
	return perms
}
