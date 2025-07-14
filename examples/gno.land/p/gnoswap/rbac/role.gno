package rbac

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// PermissionChecker is a function type
// that validates permissions for a caller.
type PermissionChecker func(caller std.Address) error

// Role represents a role information structure.
type Role struct {
	// name represents the role's identifier
	name string
	// permissions maps permission names to their respective checker functions
	permissions map[string]PermissionChecker
	addresses   map[std.Address]struct{}
}

// NewRole creates a new `Role` instance with given name.
func NewRole(roleName string) *Role {
	return &Role{
		name:        roleName,
		permissions: make(map[string]PermissionChecker),
		addresses:   make(map[std.Address]struct{}),
	}
}

func (r *Role) Name() string { return r.name }

// AddPermission adds a given permission name and checker to the role.
func (r *Role) AddPermission(permissionName string, checker PermissionChecker) {
	r.permissions[permissionName] = checker
}

// HasPermission checks if the given permission name exists
// in the role.
func (r *Role) HasPermission(permissionName string) bool {
	_, exists := r.permissions[permissionName]
	return exists
}

// RoleOption is a function type for applying options during role creation.
type RoleOption func(r *Role)

// WithPermission returns an option that registers a permission name and checker
// during role creation.
func WithPermission(permissionName string, checker PermissionChecker) RoleOption {
	return func(r *Role) {
		r.AddPermission(permissionName, checker)
	}
}

// DeclareRole registers a role and its permissions declaratively using the Functional Option pattern.
// Returns an error if the role already exists.
func (rb *RBAC) DeclareRole(roleName string, opts ...RoleOption) error {
	if rb.hasRole(roleName) {
		return ufmt.Errorf("role %s already exists", roleName)
	}
	role := NewRole(roleName)
	for _, opt := range opts {
		opt(role)
	}
	rb.roles[roleName] = role
	return nil
}
