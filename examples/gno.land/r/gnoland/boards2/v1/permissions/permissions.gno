package permissions

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/commondao"
	"gno.land/p/nt/commondao/exts/storage"
)

// ValidatorFunc defines a function type for permissions validators.
type ValidatorFunc func(boards.Permissions, boards.Args) error

// Permissions manages users, roles and permissions.
//
// This type is a default `gno.land/p/gnoland/boards` package `Permissions` implementation
// that handles boards users, roles and permissions using an underlying DAO. It also supports
// optionally setting validation functions to be triggered within `WithPermission()` method
// before a permissioned callback is called.
//
// No permissions validation is done by default.
//
// Users are allowed to have multiple roles at the same time by default, but permissions can
// be configured to only allow one role per user.
type Permissions struct {
	superRole      boards.Role
	dao            *commondao.CommonDAO
	validators     *avl.Tree // string(boards.Permission) -> BasicPermissionValidator
	public         *avl.Tree // string(boards.Permission) -> struct{}{}
	singleUserRole bool
}

// New creates a new permissions type.
func New(options ...Option) *Permissions {
	s := storage.NewMemberStorage()
	ps := &Permissions{
		validators: avl.NewTree(),
		public:     avl.NewTree(),
		dao:        commondao.New(commondao.WithMemberStorage(s)),
	}

	for _, apply := range options {
		apply(ps)
	}
	return ps
}

// DAO returns the underlying permissions DAO.
func (ps Permissions) DAO() *commondao.CommonDAO {
	return ps.dao
}

// ValidateFunc adds a custom permission validator function.
// If an existing permission function exists it's ovewritten by the new one.
func (ps *Permissions) ValidateFunc(p boards.Permission, fn ValidatorFunc) {
	ps.validators.Set(string(p), fn)
}

// SetPublicPermissions assigns permissions that are available to anyone.
// It removes previous public permissions and assigns the new ones.
// By default there are no public permissions.
func (ps *Permissions) SetPublicPermissions(permissions ...boards.Permission) {
	ps.public = avl.NewTree()
	for _, p := range permissions {
		ps.public.Set(string(p), struct{}{})
	}
}

// AddRole add a role with one or more assigned permissions.
// If role exists its permissions are overwritten with the new ones.
func (ps *Permissions) AddRole(r boards.Role, p boards.Permission, extra ...boards.Permission) {
	// If role is the super role it already has all permissions
	if ps.superRole == r {
		return
	}

	// Get member group for the role if it exists or otherwise create a new group
	grouping := ps.dao.Members().Grouping()
	name := string(r)
	group, found := grouping.Get(name)
	if !found {
		var err error
		group, err = grouping.Add(name)
		if err != nil {
			panic(err)
		}
	}

	// Save permissions within the member group overwritting any existing permissions
	group.SetMeta(append([]boards.Permission{p}, extra...))
}

// RoleExists checks if a role exists.
func (ps Permissions) RoleExists(r boards.Role) bool {
	return r == ps.superRole || ps.dao.Members().Grouping().Has(string(r))
}

// GetUserRoles returns the list of roles assigned to a user.
func (ps Permissions) GetUserRoles(user address) []boards.Role {
	groups := storage.GetMemberGroups(ps.dao.Members(), user)
	if groups == nil {
		return nil
	}

	roles := make([]boards.Role, len(groups))
	for i, name := range groups {
		roles[i] = boards.Role(name)
	}
	return roles
}

// HasRole checks if a user has a specific role assigned.
func (ps Permissions) HasRole(user address, r boards.Role) bool {
	name := string(r)
	group, found := ps.dao.Members().Grouping().Get(name)
	if !found {
		return false
	}
	return group.Members().Has(user)
}

// HasPermission checks if a user has a specific permission.
func (ps Permissions) HasPermission(user address, perm boards.Permission) bool {
	if ps.public.Has(string(perm)) {
		return true
	}

	groups := storage.GetMemberGroups(ps.dao.Members(), user)
	if groups == nil {
		return false
	}

	grouping := ps.dao.Members().Grouping()
	for _, name := range groups {
		role := boards.Role(name)
		if ps.superRole == role {
			return true
		}

		group, found := grouping.Get(name)
		if !found {
			continue
		}

		meta := group.GetMeta()
		for _, p := range meta.([]boards.Permission) {
			if p == perm {
				return true
			}
		}
	}
	return false
}

// SetUserRoles adds a new user when it doesn't exist and sets its roles.
// Method can also be called to change the roles of an existing user.
// It removes any existing user roles before assigning new ones.
// All user's roles can be removed by calling this method without roles.
func (ps *Permissions) SetUserRoles(_ realm, user address, roles ...boards.Role) {
	if len(roles) > 1 && ps.singleUserRole {
		panic("user can only have one role")
	}

	groups := storage.GetMemberGroups(ps.dao.Members(), user)
	isGuest := len(roles) == 0

	// Clear current user roles
	grouping := ps.dao.Members().Grouping()
	for _, name := range groups {
		group, found := grouping.Get(name)
		if !found {
			continue
		}

		group.Members().Remove(user)
	}

	// Add user to the storage as guest when no roles are assigned
	if isGuest {
		ps.dao.Members().Add(user)
		return
	}

	// Add user to role groups
	for _, r := range roles {
		name := string(r)
		group, found := grouping.Get(name)
		if !found {
			panic("invalid role: " + name)
		}

		group.Members().Add(user)
	}
}

// RemoveUser removes a user from permissions.
func (ps *Permissions) RemoveUser(_ realm, user address) bool {
	groups := storage.GetMemberGroups(ps.dao.Members(), user)
	if groups == nil {
		return ps.dao.Members().Remove(user)
	}

	grouping := ps.dao.Members().Grouping()
	for _, name := range groups {
		group, found := grouping.Get(name)
		if !found {
			continue
		}

		group.Members().Remove(user)
	}
	return true
}

// HasUser checks if a user exists.
func (ps Permissions) HasUser(user address) bool {
	return ps.dao.Members().Has(user)
}

// UsersCount returns the total number of users the permissioner contains.
func (ps Permissions) UsersCount() int {
	return ps.dao.Members().Size()
}

// IterateUsers iterates permissions' users.
func (ps Permissions) IterateUsers(start, count int, fn boards.UsersIterFn) (stopped bool) {
	ps.dao.Members().IterateByOffset(start, count, func(addr address) bool {
		user := boards.User{Address: addr}
		groups := storage.GetMemberGroups(ps.dao.Members(), addr)
		if groups != nil {
			user.Roles = make([]boards.Role, len(groups))
			for i, name := range groups {
				user.Roles[i] = boards.Role(name)
			}
		}

		return fn(user)
	})
	return
}

// WithPermission calls a callback when a user has a specific permission.
// It panics on error or when a permission validator fails.
// Callbacks are by default called when there is no validator registered for the permission.
// If a permission validation function exists it's called before calling the callback.
func (ps *Permissions) WithPermission(_ realm, user address, p boards.Permission, args boards.Args, cb func(realm)) {
	if !ps.HasPermission(user, p) {
		panic("unauthorized")
	}

	// Execute custom validation before calling the callback
	v, found := ps.validators.Get(string(p))
	if found {
		err := v.(ValidatorFunc)(ps, args)
		if err != nil {
			panic(err)
		}
	}

	cb(cross)
}
