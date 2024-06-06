package accesscontrol

import (
	"std"

	"gno.land/p/demo/avl"
)

// RoleData struct to store role information
type RoleData struct {
	Holder   *avl.Tree // -> std.Address -> bool
	AdminRole std.Address
}

// NewRoleData creates a new instance of RoleData
func NewRoleData(adminRole std.Address) *RoleData {
	return &RoleData{
		Holder:   avl.NewTree(),
		AdminRole: adminRole,
	}
}

// Method to check if the caller has the admin role
func (roleData *RoleData) OnlyAdmin() {
	caller := std.GetOrigCaller()
	if roleData.AdminRole != caller {
		panic("accesscontrol: caller does not have the admin role")
	}
}

// Method to check if an account has a specific role
func (roleData *RoleData) HasRole(account std.Address) bool {
	return roleData.Holder.Has(account.String())
}

// Method to grant a role to an account
func (roleData *RoleData) GrantRole(account std.Address) {
	roleData.OnlyAdmin()
	roleData.Holder.Set(account.String(), true)
	roleData.emitRoleGranted(account, std.GetOrigCaller())
}

// Method to revoke a role from an account
func (roleData *RoleData) RevokeRole(account std.Address) {
	roleData.OnlyAdmin()
	roleData.Holder.Remove(account.String())
	roleData.emitRoleRevoked(account, std.GetOrigCaller())
}

// Method to renounce a role with caller confirmation
func (roleData *RoleData) RenounceRole(callerConfirmation std.Address) {
	caller := std.GetOrigCaller()
	if callerConfirmation != caller {
		panic("accesscontrol: caller confirmation does not match account")
	}
	roleData.Holder.Remove(caller.String())
	roleData.emitRoleRevoked(caller, caller)
}

// Method to get the admin role of a specific role
func (roleData *RoleData) GetRoleAdmin() std.Address {
	return roleData.AdminRole
}

// Method to set the admin role for a specific role
func (roleData *RoleData) SetRoleAdmin(adminRole std.Address) {
	roleData.OnlyAdmin()
	previousAdminRole := roleData.AdminRole
	roleData.AdminRole = adminRole
	roleData.emitRoleAdminChanged(previousAdminRole, adminRole)
}

// Event for role granted
func (roleData *RoleData) emitRoleGranted(account std.Address, sender std.Address) {
	std.Emit("RoleGranted", "role", roleData.AdminRole.String(), "account", account.String(), "sender", sender.String())
}

// Event for role revoked
func (roleData *RoleData) emitRoleRevoked(account std.Address, sender std.Address) {
	std.Emit("RoleRevoked", "role", roleData.AdminRole.String(), "account", account.String(), "sender", sender.String())
}

// Event for role admin changed
func (roleData *RoleData) emitRoleAdminChanged(previousAdminRole std.Address, newAdminRole std.Address) {
	std.Emit("RoleAdminChanged", "previousAdminRole", previousAdminRole.String(), "newAdminRole", newAdminRole.String())
}