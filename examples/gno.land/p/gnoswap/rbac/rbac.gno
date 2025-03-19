package rbac

import (
	"std"

	"gno.land/p/demo/ufmt"
)

// RBAC encapsulates and manages toles and their permissions.
type RBAC struct {
	ownable *Ownable2Step
	// roles maps role names to their respective `Role` objects
	roles map[string]*Role
}

// New creates a new RBAC instance.
func New() *RBAC {
	return &RBAC{
		ownable: NewOwnable2Step(),
		roles:   make(map[string]*Role),
	}
}

func NewRBACWithAddress(addr std.Address) *RBAC {
	return &RBAC{
		ownable: NewOwnable2StepWithAddress(addr),
		roles:   make(map[string]*Role),
	}
}

func (rb *RBAC) Owner() std.Address { return rb.ownable.Owner() }

// PendingOwner returns the pending owner address
func (rb *RBAC) PendingOwner() std.Address { return rb.ownable.PendingOwner() }

// AcceptOwnership completes the ownership transfer process
func (rb *RBAC) AcceptOwnership() error { return rb.ownable.AcceptOwnership() }

// DropOwnership removes the owner, effectively disabling owner-only actions
func (rb *RBAC) DropOwnership() error { return rb.ownable.DropOwnership() }

// HasRole checks if an address has a specific role
func (rb *RBAC) HasRole(roleName string, addr std.Address) bool {
	role, exists := rb.roles[roleName]
	if !exists {
		return false
	}
	_, exists = role.addresses[addr]
	return exists
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

func (rb *RBAC) RemoveRole(roleName string) error {
	if !rb.hasRole(roleName) {
		return ufmt.Errorf("role %s does not exist", roleName)
	}

	if isSystemRole(roleName) {
		return ufmt.Errorf("cannot remove system role %s", roleName)
	}

	role := rb.roles[roleName]
	for permName := range role.permissions {
		if err := rb.RemovePermission(roleName, permName); err != nil {
			return ufmt.Errorf("failed to remove permission %s from role %s: %v", permName, roleName, err)
		}
	}

	delete(rb.roles, roleName)
	return nil
}

var systemRoles = map[string]bool{
	"admin":          true,
	"community_pool": true,
	"devops":         true,
	"governance":     true,
	"gov_staker":     true,
	"router":         true,
	"pool":           true,
	"position":       true,
	"staker":         true,
	"launchpad":      true,
	"emission":       true,
	"protocol_fee":   true,
	"xgns":           true,
}

func GetSystemRoles() map[string]bool {
	return systemRoles
}

// isSystemRole checks if the role is a system-critical role
// role defined in contract/r/gnoswap/access/consts.gno
func isSystemRole(roleName string) bool {
	return systemRoles[roleName]
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

// GrantRoleAndSetAddressByOwner assigns a role to an address
func (rb *RBAC) GrantRoleAndSetAddressByOwner(roleName string, addr std.Address) error {
	if !rb.ownable.CallerIsOwner() {
		return ErrUnauthorized
	}
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("role %s does not exist", roleName)
	}
	if !addr.IsValid() {
		return ufmt.Errorf("invalid address %s", addr.String())
	}
	role.addresses[addr] = struct{}{}
	return nil
}

// RevokeRole removes a role from an address
func (rb *RBAC) RevokeRole(roleName string, addr std.Address) error {
	if !rb.ownable.CallerIsOwner() {
		return ErrUnauthorized
	}
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("role %s does not exist", roleName)
	}
	if !addr.IsValid() {
		return ufmt.Errorf("invalid address %s", addr.String())
	}
	delete(role.addresses, addr)
	return nil
}

// ChangeOwner2Step initiates the two-step ownership transfer process
func (rb *RBAC) ChangeOwner2Step(newOwner std.Address) error {
	return rb.ownable.TransferOwnership(newOwner)
}

// GetAddressesByRole returns all addresses with a specific role
func (rb *RBAC) GetAddressesByRole(roleName string) ([]std.Address, error) {
	role, exists := rb.roles[roleName]
	if !exists {
		return nil, ufmt.Errorf("role %s does not exist", roleName)
	}

	addrs := make([]std.Address, 0, len(role.addresses))
	for addr := range role.addresses {
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

// RemoveAddressFromRole removes a specific address from a role
func (rb *RBAC) RemoveAddressFromRole(roleName string, addr std.Address) error {
	if !rb.ownable.CallerIsOwner() {
		return ErrUnauthorized
	}
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("role %s does not exist", roleName)
	}

	_, hasAddr := role.addresses[addr]
	if !hasAddr {
		return ufmt.Errorf("address %s does not have role %s", addr.String(), roleName)
	}
	delete(role.addresses, addr)
	return nil
}

func (rb *RBAC) RemoveAllAddressesFromRole(roleName string) error {
	if !rb.ownable.CallerIsOwner() {
		return ErrUnauthorized
	}
	role, exists := rb.roles[roleName]
	if !exists {
		return ufmt.Errorf("role %s does not exist", roleName)
	}
	// Create a new empty map to replace the current one
	role.addresses = make(map[std.Address]struct{})
	return nil
}
