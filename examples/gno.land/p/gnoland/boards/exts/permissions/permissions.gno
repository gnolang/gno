package permissions

import (
	"strings"

	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/groups/v0"
)

// ValidatorFunc defines a function type for permissions validators.
//
// SECURITY: validators run inside WithPermission holding the live
// Permissions value, with full mutation access (SetUserRoles, RemoveUser,
// AddRole, ...) under the owning realm's authority. Register only functions
// the owning realm controls, and never expose ValidateFunc or the
// *Permissions value across a realm boundary.
type ValidatorFunc func(boards.Permissions, boards.Args) error

// Permissions manages users, roles and permissions.
//
// This type is a default `gno.land/p/gnoland/boards` package `Permissions`
// implementation that handles boards users, roles and permissions using an
// underlying groups.Group: the base set holds every user (guests included),
// and each boards role is a group Role whose member set is kept a subset of
// the base set, with the role's boards.PermissionSet stored in the role meta
// (a value type, per the groups meta rule). It also supports optionally
// setting validation functions to be triggered within `WithPermission()`
// method before a permissioned callback is called.
//
// No permissions validation is done by default.
//
// Users are allowed to have multiple roles at the same time by default, but
// permissions can be configured to only allow one role per user.
type Permissions struct {
	superRole      boards.Role
	group          *groups.Group
	public         boards.PermissionSet
	validators     *bptree.BPTree // string(boards.Permission) -> ValidatorFunc
	singleUserRole bool
}

// New creates a new permissions type.
func New(options ...Option) *Permissions {
	ps := &Permissions{
		validators: bptree.NewBPTree32(),
		group:      groups.NewGroup(),
	}

	for _, apply := range options {
		apply(ps)
	}
	return ps
}

// ValidateFunc adds a custom permission validator function.
// If an existing permission function exists it's overwritten by the new one.
func (ps *Permissions) ValidateFunc(p boards.Permission, fn ValidatorFunc) {
	ps.validators.Set(p.String(), fn)
}

// SetPublicPermissions assigns permissions that are available to anyone.
// It removes previous public permissions and assigns the new ones.
// By default there are no public permissions.
func (ps *Permissions) SetPublicPermissions(permissions ...boards.Permission) {
	ps.public = boards.NewPermissionSet(permissions...)
}

// AddRole adds a role with one or more assigned permissions.
// If role exists its permissions are overwritten with the new ones.
func (ps *Permissions) AddRole(r boards.Role, p boards.Permission, extra ...boards.Permission) {
	name := string(r)
	if strings.TrimSpace(name) == "" {
		panic("role name is required")
	}

	// If role is the super role it already has all permissions
	if ps.superRole == r {
		return
	}

	// Get the role if it exists or otherwise register a new one
	role, found := ps.group.GetRole(name)
	if !found {
		var err error
		role, err = ps.group.AddRole(name)
		if err != nil {
			panic(err)
		}
	}

	// Save permissions within the role meta overwriting any existing permissions
	permissions := append([]boards.Permission{p}, extra...)
	role.SetMeta(boards.NewPermissionSet(permissions...))
}

// RoleExists checks if a role exists.
func (ps Permissions) RoleExists(r boards.Role) bool {
	return r == ps.superRole || ps.group.HasRole(string(r))
}

// GetUserRoles returns the list of roles assigned to a user.
func (ps Permissions) GetUserRoles(user address) []boards.Role {
	names := ps.group.RolesContaining(user)
	if names == nil {
		return nil
	}

	roles := make([]boards.Role, len(names))
	for i, name := range names {
		roles[i] = boards.Role(name)
	}
	return roles
}

// HasRole checks if a user has a specific role assigned.
func (ps Permissions) HasRole(user address, r boards.Role) bool {
	role, found := ps.group.GetRole(string(r))
	if !found {
		return false
	}
	return role.Members().Has(user)
}

// HasPermission checks if a user has a specific permission.
func (ps Permissions) HasPermission(user address, perm boards.Permission) bool {
	if ps.public.Has(perm) {
		return true
	}

	for _, name := range ps.group.RolesContaining(user) {
		if ps.superRole == boards.Role(name) {
			return true
		}

		role, found := ps.group.GetRole(name)
		if !found {
			continue
		}

		if perms, ok := role.Meta().(boards.PermissionSet); ok && perms.Has(perm) {
			return true
		}
	}
	return false
}

// SetUserRoles adds a new user when it doesn't exist and sets its roles.
// Method can also be called to change the roles of an existing user.
// It removes any existing user roles before assigning new ones.
// All user's roles can be removed by calling this method without roles.
func (ps *Permissions) SetUserRoles(user address, roles ...boards.Role) {
	if len(roles) > 1 && ps.singleUserRole {
		panic("user can only have one role")
	}

	// Resolve every role name upfront so an invalid name panics before any
	// state is mutated.
	newRoles := make([]*groups.Role, len(roles))
	for i, r := range roles {
		role, found := ps.group.GetRole(string(r))
		if !found {
			panic("invalid role: " + string(r))
		}
		newRoles[i] = role
	}

	// Clear current user roles
	for _, name := range ps.group.RolesContaining(user) {
		if role, found := ps.group.GetRole(name); found {
			role.Members().Remove(user)
		}
	}

	// Every user is a base set member, with or without roles; role member
	// sets are kept subsets of the base set.
	ps.group.Add(user)

	// Add user to role member sets
	for _, role := range newRoles {
		role.Members().Add(user)
	}
}

// RemoveUser removes a user from permissions.
func (ps *Permissions) RemoveUser(user address) bool {
	return ps.group.RemoveFromAll(user)
}

// HasUser checks if a user exists.
func (ps Permissions) HasUser(user address) bool {
	return ps.group.Has(user)
}

// UsersCount returns the total number of users the permissioner contains.
func (ps Permissions) UsersCount() int {
	return ps.group.Size()
}

// IterateUsers iterates permissions' users.
func (ps Permissions) IterateUsers(start, count int, fn boards.UsersIterFn) (stopped bool) {
	return ps.group.Iterate(start, count, func(addr address) bool {
		return fn(boards.User{
			Address: addr,
			Roles:   ps.GetUserRoles(addr),
		})
	})
}

// WithPermission calls a callback when a user has a specific permission.
// It panics on error or when a permission validator fails.
// Callbacks are by default called when there is no validator function registered for the permission.
// If a permission validation function exists it's called before calling the callback.
func (ps *Permissions) WithPermission(user address, p boards.Permission, args boards.Args, cb func()) {
	if !ps.HasPermission(user, p) {
		panic("unauthorized, user " + user.String() + " doesn't have the required permission")
	}

	// Execute custom validation before calling the callback
	if v := ps.validators.Get(p.String()); v != nil {
		err := v.(ValidatorFunc)(ps, args)
		if err != nil {
			panic(err)
		}
	}

	cb()
}
