package rbac

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// RBAC encapsulates and manages toles and their permissions.
type RBAC struct {
	// roles maps role names to their respective `Role` objects
	roles map[string]*Role
}

// New creates a new RBAC instance.
func New() *RBAC {
	return &RBAC{
		roles: make(map[string]*Role),
	}
}

func (rb *RBAC) hasRole(name string) bool {
	_, exists := rb.roles[name]
	return exists
}

// RegisterRole registers a role with the given role name.
// Returns an error if the role already exists.
func (rb *RBAC) RegisterRole(roleName string) error {
	if rb.hasRole(roleName) {
		return ufmt.Errorf("role %s already exists", roleName)
	}
	rb.roles[roleName] = NewRole(roleName)
	return nil
}

// RegisterPermission registers a permission name and checker
// for the specific role.
func (rb *RBAC) RegisterPermission(
	roleName, permissionName string,
	checker PermissionChecker,
) error {
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("role %s does not exist", roleName)
	}
	role.AddPermission(permissionName, checker)
	return nil
}

// CheckPermission verifies if the caller has the specific permission.
func (rb *RBAC) CheckPermission(
	roleName, permissionName string,
	caller std.Address,
) error {
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("role %s does not exist", roleName)
	}
	checker, exists := role.permissions[permissionName]
	if !exists {
		return ufmt.Errorf("permission %s does not exist for role %s", permissionName, roleName)
	}
	return checker(caller)
}

// UpdatePermission updates the checker for a specific permission
// in a role.
func (rb *RBAC) UpdatePermission(
	roleName, permissionName string,
	newChecker PermissionChecker,
) error {
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("role %s does not exist", roleName)
	}
	if !role.HasPermission(permissionName) {
		return ufmt.Errorf("permission %s does not exist for role %s", permissionName, roleName)
	}
	role.AddPermission(permissionName, newChecker)
	return nil
}

// RemovePermission removes a permission from a role.
func (rb *RBAC) RemovePermission(
	roleName, permissionName string,
) error {
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("cannot remove permission from non-existent role %s", roleName)
	}
	if !role.HasPermission(permissionName) {
		return ufmt.Errorf("permission %s does not exist for role %s", permissionName, roleName)
	}
	delete(role.permissions, permissionName)
	return nil
}
