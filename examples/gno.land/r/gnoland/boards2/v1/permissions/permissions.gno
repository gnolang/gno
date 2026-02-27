package permissions

import (
	"gno.land/p/gnoland/boards"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/commondao"
	"gno.land/p/nt/commondao/exts/storage"
)

// ValidatorFunc defines a function type for permissions validators.
type ValidatorFunc func(boards.Permissions, boards.Args) error

// Permisions manages users, roles and permissions.
//
// This type is a default `gno.land/p/gnoland/boards` package `Permissions` implementation
// that handles boards users, roles and permissions using an underlying DAO. It also supports
// optionally setting validation functions to be triggered within `WithPermission()` method
// before a permissioned callback is called.
//
// No permissions validation is done by default.
type Permisions struct {
	superRole  boards.Role
	dao        *commondao.CommonDAO
	validators *avl.Tree // string(boards.Permission) -> BasicPermissionValidator
	public     *avl.Tree // string(boards.Permission) -> struct{}{}
}

// New creates a new permissions type.
func New() *Permisions {
	s := storage.NewMemberStorage()
	return &Permisions{
		validators: avl.NewTree(),
		public:     avl.NewTree(),
		dao:        commondao.New(commondao.WithMemberStorage(s)),
	}
}

// DAO returns the underlying permissions DAO.
func (bp Permisions) DAO() *commondao.CommonDAO {
	return bp.dao
}

// ValidateFunc adds a custom permission validator function.
// If an existing permission function exists it's ovewritten by the new one.
func (bp *Permisions) ValidateFunc(p boards.Permission, fn ValidatorFunc) {
	bp.validators.Set(string(p), fn)
}

// SetPublicPermissions assigns permissions that are available to anyone.
// It removes previous public permissions and assigns the new ones.
// By default there are no public permissions.
func (bp *Permisions) SetPublicPermissions(permissions ...boards.Permission) {
	bp.public = avl.NewTree()
	for _, p := range permissions {
		bp.public.Set(string(p), struct{}{})
	}
}

// SetSuperRole assigns a super role.
// A super role is one that have all permissions.
// These type of role doesn't need to be mapped to any permission.
func (bp *Permisions) SetSuperRole(r boards.Role) {
	if bp.superRole == r {
		return
	}

	name := string(r)
	bp.dao.Members().Grouping().Add(name)
	bp.superRole = r
}

// AddRole add a role with one or more assigned permissions.
// If role exists its permissions are overwritten with the new ones.
func (bp *Permisions) AddRole(r boards.Role, p boards.Permission, extra ...boards.Permission) {
	// Get member group for the role if it exists or otherwise create a new group
	grouping := bp.dao.Members().Grouping()
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
func (bp Permisions) RoleExists(r boards.Role) bool {
	return (bp.superRole != "" && r == bp.superRole) || bp.dao.Members().Grouping().Has(string(r))
}

// GetUserRoles returns the list of roles assigned to a user.
func (bp Permisions) GetUserRoles(user address) []boards.Role {
	groups := storage.GetMemberGroups(bp.dao.Members(), user)
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
func (bp Permisions) HasRole(user address, r boards.Role) bool {
	name := string(r)
	group, found := bp.dao.Members().Grouping().Get(name)
	if !found {
		return false
	}
	return group.Members().Has(user)
}

// HasPermission checks if a user has a specific permission.
func (bp Permisions) HasPermission(user address, perm boards.Permission) bool {
	if bp.public.Has(string(perm)) {
		return true
	}

	groups := storage.GetMemberGroups(bp.dao.Members(), user)
	if groups == nil {
		return false
	}

	grouping := bp.dao.Members().Grouping()
	for _, name := range groups {
		role := boards.Role(name)
		if bp.superRole == role {
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
// All user's roles can be removed by calling this method without roles.
func (bp *Permisions) SetUserRoles(_ realm, user address, roles ...boards.Role) {
	groups := storage.GetMemberGroups(bp.dao.Members(), user)
	isGuest := len(roles) == 0

	// If user has roles remove it from the groups its currently assigned
	grouping := bp.dao.Members().Grouping()
	if isGuest && groups != nil {
		for _, name := range groups {
			group, found := grouping.Get(name)
			if !found {
				continue
			}

			group.Members().Remove(user)
		}
	}

	// Add user to the storage as guest when no roles are assigned
	if isGuest {
		bp.dao.Members().Add(user)
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
func (bp *Permisions) RemoveUser(_ realm, user address) bool {
	groups := storage.GetMemberGroups(bp.dao.Members(), user)
	if groups == nil {
		return bp.dao.Members().Remove(user)
	}

	grouping := bp.dao.Members().Grouping()
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
func (bp Permisions) HasUser(user address) bool {
	return bp.dao.Members().Has(user)
}

// UsersCount returns the total number of users the permissioner contains.
func (bp Permisions) UsersCount() int {
	return bp.dao.Members().Size()
}

// IterateUsers iterates permissions' users.
func (bp Permisions) IterateUsers(start, count int, fn boards.UsersIterFn) (stopped bool) {
	bp.dao.Members().IterateByOffset(start, count, func(addr address) bool {
		user := boards.User{Address: addr}
		groups := storage.GetMemberGroups(bp.dao.Members(), addr)
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
func (bp *Permisions) WithPermission(_ realm, user address, p boards.Permission, args boards.Args, cb func(realm)) {
	if !bp.HasPermission(user, p) {
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
