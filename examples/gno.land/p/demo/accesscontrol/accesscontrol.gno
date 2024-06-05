package accesscontrol

import (
	"std"

	"gno.land/p/demo/avl"
)

// RoleData struct to store role information
type RoleData struct {
	hasRole   *avl.Tree // -> std.Address -> bool
	adminRole std.Address
}

// NewRoleData creates a new instance of RoleData
func NewRoleData(adminRole std.Address) *RoleData {
	return &RoleData{
		hasRole:   avl.NewTree(),
		adminRole: adminRole,
	}
}

// Method to check if the caller has the admin role
func (roleData *RoleData) OnlyAdmin() {
	caller := std.GetOrigCaller()
	if roleData.adminRole != caller {
		panic("accesscontrol: caller does not have the admin role")
	}
}

// Method to check if an account has a specific role
func (roleData *RoleData) HasRole(account std.Address) bool {
	return roleData.hasRole.Has(account.String())
}

// Method to grant a role to an account
func (roleData *RoleData) GrantRole(account std.Address) {
	roleData.OnlyAdmin()
	roleData.hasRole.Set(account.String(), true)
	roleData.emitRoleGranted(account, std.GetOrigCaller())
}

// Method to revoke a role from an account
func (roleData *RoleData) RevokeRole(account std.Address) {
	roleData.OnlyAdmin()
	roleData.hasRole.Remove(account.String())
	roleData.emitRoleRevoked(account, std.GetOrigCaller())
}

// Method to renounce a role with caller confirmation
func (roleData *RoleData) RenounceRole(callerConfirmation std.Address) {
	caller := std.GetOrigCaller()
	if callerConfirmation != caller {
		panic("accesscontrol: caller confirmation does not match account")
	}
	roleData.hasRole.Remove(caller.String())
	roleData.emitRoleRevoked(caller, caller)
}

// Method to get the admin role of a specific role
func (roleData *RoleData) GetRoleAdmin() std.Address {
	return roleData.adminRole
}

// Method to set the admin role for a specific role
func (roleData *RoleData) SetRoleAdmin(adminRole std.Address) {
	roleData.OnlyAdmin()
	previousAdminRole := roleData.adminRole
	roleData.adminRole = adminRole
	roleData.emitRoleAdminChanged(previousAdminRole, adminRole)
}

// Event for role granted
func (roleData *RoleData) emitRoleGranted(account std.Address, sender std.Address) {
	std.Emit("RoleGranted", "role", roleData.adminRole.String(), "account", account.String(), "sender", sender.String())
}

// Event for role revoked
func (roleData *RoleData) emitRoleRevoked(account std.Address, sender std.Address) {
	std.Emit("RoleRevoked", "role", roleData.adminRole.String(), "account", account.String(), "sender", sender.String())
}

// Event for role admin changed
func (roleData *RoleData) emitRoleAdminChanged(previousAdminRole std.Address, newAdminRole std.Address) {
	std.Emit("RoleAdminChanged", "previousAdminRole", previousAdminRole.String(), "newAdminRole", newAdminRole.String())
}

/*package accesscontrol

import (
    "gno.land/p/demo/avl"
    "std"
)

// RoleData struct to store role information
type RoleData struct {
    HasRole   *avl.Tree // -> std.Address -> bool
    AdminRole std.Address 
}

// NewRoleData creates a new instance of RoleData
func NewRoleData(adminRole std.Address) *RoleData {
    return &RoleData{
        HasRole:   avl.NewTree(),
        AdminRole: adminRole,
    }
}

// Modifier to check if the caller has the admin role
func OnlyAdmin(roleData *RoleData) {
    caller := std.GetOrigCaller()
    if roleData.AdminRole != caller {
        panic("accesscontrol: caller does not have the admin role")
    }
}

// Function to check if an account has a specific role
func HasRole(roleData *RoleData, account std.Address) bool {
    return roleData.HasRole.Has(account.String())
}

// Function to grant a role to an account
func GrantRole(roleData *RoleData, account std.Address) {
    OnlyAdmin(roleData)
    roleData.HasRole.Set(account.String(), true)
    emitRoleGranted(roleData, account, std.GetOrigCaller())
    std.Emit("RoleGranted", "role", roleData, "account", account.String(), "sender", sender.String())
}

// Function to revoke a role from an account
func RevokeRole(roleData *RoleData, account std.Address) {
    OnlyAdmin(roleData)
    roleData.HasRole.Remove(account.String())
    emitRoleRevoked(roleData, account, std.GetOrigCaller())
}

// Function to renounce a role with caller confirmation
func RenounceRole(roleData *RoleData, callerConfirmation std.Address) {
    caller := std.GetOrigCaller()
    if callerConfirmation != caller {
        panic("accesscontrol: caller confirmation does not match account")
    }
    roleData.HasRole.Remove(caller.String())
    emitRoleRevoked(roleData, caller, caller)
}

// Function to get the admin role of a specific role
func GetRoleAdmin(roleData *RoleData) std.Address {
    return roleData.AdminRole
}

// Function to set the admin role for a specific role
func SetRoleAdmin(roleData *RoleData, adminRole std.Address) {
    OnlyAdmin(roleData)
    previousAdminRole := roleData.AdminRole
    roleData.AdminRole = adminRole
    emitRoleAdminChanged(roleData, previousAdminRole, adminRole)
}


// Event for role granted
func emitRoleGranted(roleData *RoleData, account std.Address, sender std.Address) {
    std.Emit("RoleGranted", "role", roleData.AdminRole.String(), "account", account.String(), "sender", sender.String())
}

// Event for role revoked
func emitRoleRevoked(roleData *RoleData, account std.Address, sender std.Address) {
    std.Emit("RoleRevoked", "role", roleData.AdminRole.String(), "account", account.String(), "sender", sender.String())
}

// Event for role admin changed
func emitRoleAdminChanged(roleData *RoleData, previousAdminRole std.Address, newAdminRole std.Address) {
    std.Emit("RoleAdminChanged", "previousAdminRole", previousAdminRole.String(), "newAdminRole", newAdminRole.String())
}*/