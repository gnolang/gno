package boards2

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/nt/commondao"

	"gno.land/r/demo/users"
)

// DefaultPermissions manages users, roles and permissions.
type DefaultPermissions struct {
	superRole Role
	dao       *commondao.CommonDAO
	users     *avl.Tree // string(std.Address) -> []Role
	roles     *avl.Tree // string(role) -> []Permission
}

// NewDefaultPermissions creates a new permissions type.
// This type is a default implementation to handle users, roles and permissions.
func NewDefaultPermissions(dao *commondao.CommonDAO) *DefaultPermissions {
	if dao == nil {
		panic("default permissions require a DAO")
	}

	return &DefaultPermissions{
		dao:   dao,
		roles: avl.NewTree(),
		users: avl.NewTree(),
	}
}

// SetSuperRole assigns a super role.
// A super role is one that have all permissions.
// These type of role doesn't need to be mapped to any permission.
func (dp *DefaultPermissions) SetSuperRole(r Role) {
	dp.superRole = r
}

// AddRole add a role with one or more assigned permissions.
func (dp *DefaultPermissions) AddRole(r Role, p Permission, extra ...Permission) {
	dp.roles.Set(string(r), append([]Permission{p}, extra...))
}

// RoleExists checks if a role exists.
func (dp DefaultPermissions) RoleExists(r Role) bool {
	if dp.superRole != "" && r == dp.superRole {
		return true
	}

	return dp.roles.Iterate("", "", func(name string, _ interface{}) bool {
		return Role(name) == r
	})
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
	for _, r := range dp.GetUserRoles(user) {
		if dp.superRole == r {
			return true
		}

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

// AddUser adds a new user to permissions.
func (dp *DefaultPermissions) AddUser(user std.Address, roles ...Role) error {
	if dp.users.Has(user.String()) {
		return errors.New("user already exists")
	}

	if err := dp.dao.AddMember(user); err != nil {
		return err
	}
	return dp.setUserRoles(user, roles...)
}

// SetUserRoles sets the roles of a user.
func (dp *DefaultPermissions) SetUserRoles(user std.Address, roles ...Role) error {
	if !dp.users.Has(user.String()) {
		return errors.New("user not found")
	}
	return dp.setUserRoles(user, roles...)
}

// RemoveUser removes a user from permissions.
func (dp *DefaultPermissions) RemoveUser(user std.Address) bool {
	_, removed := dp.users.Remove(user.String())
	dp.dao.RemoveMember(user)
	return removed
}

// HasUser checks if a user exists.
func (dp DefaultPermissions) HasUser(user std.Address) bool {
	return dp.users.Has(user.String())
}

// WithPermission calls a callback when a user has a specific permission.
// It panics on error or when a handler panics.
// Callbacks are by default called when there is no handle registered for the permission.
func (dp *DefaultPermissions) WithPermission(user std.Address, perm Permission, args Args, cb func(Args)) {
	if !dp.HasPermission(user, perm) || !dp.dao.IsMember(user) {
		panic("unauthorized")
	}

	switch perm {
	case PermissionBoardCreate:
		dp.handleBoardCreate(args, cb)
	case PermissionBoardRename:
		dp.handleBoardRename(args, cb)
	case PermissionMemberInvite:
		dp.handleMemberInvite(args, cb)
	case PermissionRoleChange:
		dp.handleRoleChange(args, cb)
	default:
		cb(args)
	}
}

func (dp *DefaultPermissions) setUserRoles(user std.Address, roles ...Role) error {
	for _, r := range roles {
		if !dp.RoleExists(r) {
			return errors.New("invalid role: " + string(r))
		}
	}

	dp.users.Set(user.String(), append([]Role(nil), roles...))
	return nil
}

func (DefaultPermissions) handleBoardCreate(args Args, cb func(Args)) {
	name, ok := args[0].(string)
	if !ok {
		panic("expected board name to be a string")
	}

	assertValidBoardNameLength(name)
	assertBoardNameIsNotAddress(name)
	assertBoardNameBelongsToCaller(name)

	cb(args)
}

func (DefaultPermissions) handleBoardRename(args Args, cb func(Args)) {
	newName, ok := args[2].(string)
	if !ok {
		panic("expected new board name to be a string")
	}

	assertValidBoardNameLength(newName)
	assertBoardNameIsNotAddress(newName)
	assertBoardNameBelongsToCaller(newName)

	cb(args)
}

func (dp DefaultPermissions) handleMemberInvite(args Args, cb func(Args)) {
	// Make sure that only owners invite other owners
	role, ok := args[1].(Role)
	if !ok {
		panic("expected a valid new member role")
	}

	if role == RoleOwner {
		if !dp.HasRole(std.OriginCaller(), RoleOwner) {
			panic("only owners are allowed to invite other owners")
		}
	}

	cb(args)
}

func (dp DefaultPermissions) handleRoleChange(args Args, cb func(Args)) {
	// Owners and Admins can change roles.
	// Admins should not be able to assign or remove the Owner role from members.
	if dp.HasRole(std.OriginCaller(), RoleAdmin) {
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

			if dp.HasRole(member, RoleOwner) {
				panic("admins are not allowed to remove the Owner role")
			}
		}
	}

	cb(args)
}

func createDefaultPermissions(owner std.Address) *DefaultPermissions {
	perms := NewDefaultPermissions(commondao.New())
	perms.SetSuperRole(RoleOwner)
	perms.AddRole(RoleAdmin, PermissionBoardCreate)
	perms.AddUser(owner, RoleOwner)
	return perms
}

func assertBoardNameIsNotAddress(s string) {
	if std.Address(s).IsValid() {
		panic("addresses are not allowed as board name")
	}
}

func assertValidBoardNameLength(name string) {
	if len(name) < 6 {
		panic("the minimum allowed board name length is 6 characters")
	}
}

func assertBoardNameBelongsToCaller(name string) {
	// When the board name is the name of a registered user
	// check that caller is the owner of the name.
	user := users.GetUserByName(name)
	if user != nil && user.Address != std.OriginCaller() {
		panic("board name is a user name registered to a different user")
	}
}
