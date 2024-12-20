package boards2

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/boards2/admindao"
)

type (
	// PermissionsHandlerFunc defines a function to handle permission callbacks.
	// Handlers are called by the `WithPermission()` method to execute callbacks
	// when users have the permission assigned.
	PermissionsHandlerFunc func(Permissions, Args, func(Args))

	// DefaultPermissions manages users, roles and permissions.
	DefaultPermissions struct {
		superRole Role
		dao       *admindao.AdminDAO
		handlers  *avl.Tree // string(permission) -> PermissionsHandlerFunc
		users     *avl.Tree // string(std.Address) -> []Role
		roles     *avl.Tree // string(role) -> []Permission
	}
)

// NewDefaultPermissions creates a new permissions type.
// This type is a default implementation to handle users, roles and permissions.
func NewDefaultPermissions(dao *admindao.AdminDAO, options ...DefaultPermissionsOption) *DefaultPermissions {
	dp := &DefaultPermissions{
		dao:      dao,
		handlers: avl.NewTree(),
		roles:    avl.NewTree(),
		users:    avl.NewTree(),
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

// RoleExists checks if a role exists.
func (dp DefaultPermissions) RoleExists(r Role) bool {
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
	// TODO: Should we check that the user belongs to the DAO?
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

	for _, r := range roles {
		if !dp.RoleExists(r) {
			return errors.New("invalid role: " + string(r))
		}
	}

	if err := dp.dao.AddMember(user); err != nil {
		return err
	}

	dp.users.Set(user.String(), append([]Role(nil), roles...))
	return nil
}

// RemoveUser removes a user from permissions.
func (dp *DefaultPermissions) RemoveUser(user std.Address) bool {
	_, removed := dp.users.Remove(user.String())
	dp.dao.RemoveMember(user)
	return removed
}

// HandleFunc registers a handler function for a permission.
func (dp *DefaultPermissions) HandleFunc(p Permission, fn PermissionsHandlerFunc) {
	dp.handlers.Set(string(p), fn)
}

// WithPermission calls a callback when a user has a specific permission.
// It panics on error or when a handler panics.
// Callbacks are by default called when there is no handle registered for the permission.
func (dp *DefaultPermissions) WithPermission(user std.Address, perm Permission, args Args, cb func(Args)) {
	if !dp.HasPermission(user, perm) || !dp.dao.IsMember(user) {
		panic("unauthorized")
	}

	h, found := dp.handlers.Get(string(perm))
	if !found {
		cb(args) // TODO: Should we fail instead?
		return
	}

	fn := h.(PermissionsHandlerFunc)
	fn(dp, args, cb)
}

func createDefaultPermissions(owner std.Address) *DefaultPermissions {
	// TODO: DAO should be a different realm or proposal and voting functions should be part of boards realm?
	//       Permissions and DAO mechanics should be discussed and improved. Add `GetDAO()` to `Permissions`??
	perms := NewDefaultPermissions(
		admindao.New(admindao.WithMember(owner)),
		WithSuperRole(RoleOwner),
		WithRole(RoleAdmin, PermissionBoardCreate, PermissionMemberInvite),
		// TODO: Finish assigning all roles and permissions
		// WithRole(RoleModerator, permissions...),
		WithUser(owner, RoleOwner),
	)
	perms.HandleFunc(PermissionBoardCreate, handleBoardCreate)
	perms.HandleFunc(PermissionMemberInvite, handleMemberInvite)
	return perms
}
