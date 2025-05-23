package boards2

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/nt/commondao"

	"gno.land/r/sys/users"
)

// BasicPermissions manages users, roles and permissions.
type BasicPermissions struct {
	superRole Role
	dao       *commondao.CommonDAO
	users     *avl.Tree // string(std.Address) -> []Role
	roles     *avl.Tree // string(role) -> []Permission
}

// NewBasicPermissions creates a new permissions type.
// This type is a default implementation to handle users, roles and permissions.
func NewBasicPermissions(dao *commondao.CommonDAO) *BasicPermissions {
	if dao == nil {
		panic("basic permissions require a DAO")
	}

	return &BasicPermissions{
		dao:   dao,
		roles: avl.NewTree(),
		users: avl.NewTree(),
	}
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

// AddUser adds a new user to permissions.
func (bp *BasicPermissions) AddUser(user std.Address, roles ...Role) error {
	if bp.dao.Members().Has(user) {
		return errors.New("user already exists")
	}

	bp.dao.Members().Add(user)
	return bp.setUserRoles(user, roles...)
}

// SetUserRoles sets the roles of a user.
func (bp *BasicPermissions) SetUserRoles(user std.Address, roles ...Role) error {
	if !bp.HasUser(user) {
		return errors.New("user not found")
	}
	return bp.setUserRoles(user, roles...)
}

// RemoveUser removes a user from permissions.
func (bp *BasicPermissions) RemoveUser(user std.Address) bool {
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
func (bp *BasicPermissions) WithPermission(user std.Address, perm Permission, args Args, cb func(Args)) {
	if !bp.HasPermission(user, perm) || !bp.dao.Members().Has(user) {
		panic("unauthorized")
	}

	switch perm {
	case PermissionBoardCreate:
		bp.handleBoardCreate(args, cb)
	case PermissionBoardRename:
		bp.handleBoardRename(args, cb)
	case PermissionMemberInvite:
		bp.handleMemberInvite(args, cb)
	case PermissionRoleChange:
		bp.handleRoleChange(args, cb)
	default:
		cb(args)
	}
}

func (bp *BasicPermissions) setUserRoles(user std.Address, roles ...Role) error {
	for _, r := range roles {
		if !bp.RoleExists(r) {
			return errors.New("invalid role: " + string(r))
		}
	}

	bp.users.Set(user.String(), append([]Role(nil), roles...))
	return nil
}

func (BasicPermissions) handleBoardCreate(args Args, cb func(Args)) {
	name, ok := args[0].(string)
	if !ok {
		panic("expected board name to be a string")
	}

	assertBoardNameIsNotAddress(name)
	assertBoardNameBelongsToCaller(name)

	cb(args)
}

func (BasicPermissions) handleBoardRename(args Args, cb func(Args)) {
	newName, ok := args[2].(string)
	if !ok {
		panic("expected new board name to be a string")
	}

	assertBoardNameIsNotAddress(newName)
	assertBoardNameBelongsToCaller(newName)

	cb(args)
}

func (bp BasicPermissions) handleMemberInvite(args Args, cb func(Args)) {
	// Make sure that only owners invite other owners
	role, ok := args[1].(Role)
	if !ok {
		panic("expected a valid new member role")
	}

	if role == RoleOwner {
		if !bp.HasRole(std.PreviousRealm().Address(), RoleOwner) {
			panic("only owners are allowed to invite other owners")
		}
	}

	cb(args)
}

func (bp BasicPermissions) handleRoleChange(args Args, cb func(Args)) {
	// Owners and Admins can change roles.
	// Admins should not be able to assign or remove the Owner role from members.
	if bp.HasRole(std.PreviousRealm().Address(), RoleAdmin) {
		role, ok := args[2].(Role)
		if !ok {
			panic("expected a valid member role")
		}

		if role == RoleOwner {
			panic("admins are not allowed to promote members to Owner")
		} else {
			member, ok := args[1].(std.Address)
			if !ok {
				panic("expected a valid member address")
			}

			if bp.HasRole(member, RoleOwner) {
				panic("admins are not allowed to remove the Owner role")
			}
		}
	}

	cb(args)
}

func createBasicPermissions(owners ...std.Address) *BasicPermissions {
	perms := NewBasicPermissions(commondao.New())
	perms.SetSuperRole(RoleOwner)
	perms.AddRole(RoleAdmin, PermissionBoardCreate)
	for _, owner := range owners {
		perms.AddUser(owner, RoleOwner)
	}
	return perms
}

func assertBoardNameIsNotAddress(s string) {
	if std.Address(s).IsValid() {
		panic("addresses are not allowed as board name")
	}
}

func assertBoardNameBelongsToCaller(name string) {
	// When the board name is the name of a registered user
	// check that caller is the owner of the name.
	user, _ := users.ResolveName(name)
	if user != nil && user.Addr() != std.PreviousRealm().Address() {
		panic("board name is a user name registered to a different user")
	}
}
