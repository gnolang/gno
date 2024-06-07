package accesscontrol

import (
	"std"
	"gno.land/p/demo/avl"
)

// RoleData struct to store role information
type RoleData struct {
	Name      string
	Holder    *avl.Tree // -> std.Address -> bool
	AdminRole std.Address
}

// NewRoleData creates a new instance of RoleData
func NewRoleData(name string, adminRole std.Address) *RoleData {
	return &RoleData{
		Name:      name,
		Holder:    avl.NewTree(),
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

// Method to create a new role within the realm
func (roleData *RoleData) CreateRole(name string, adminRole std.Address) *RoleData {
	roleData.OnlyAdmin()
    std.Emit("RoleCreated", "roleName", name, "adminRole", adminRole.String(), "sender", std.GetOrigCaller().String())
    return NewRoleData(name, adminRole)
}

// Method to check if an account has a specific role
func (roleData *RoleData) HasRole(account std.Address) bool {
	return roleData.Holder.Has(account.String())
}

// Method to grant a role to an account
func (roleData *RoleData) GrantRole(account std.Address) {
	roleData.OnlyAdmin()
	roleData.Holder.Set(account.String(), true)
	std.Emit("RoleGranted", "roleName", roleData.Name, "account", account.String(), "sender", std.GetOrigCaller().String())
}

// Method to revoke a role from an account
func (roleData *RoleData) RevokeRole(account std.Address) {
	roleData.OnlyAdmin()
	roleData.Holder.Remove(account.String())
	std.Emit("RoleRevoked", "roleName", roleData.Name, "account", account.String(), "sender", std.GetOrigCaller().String())
}

// Method to renounce a role with caller confirmation
func (roleData *RoleData) RenounceRole(callerConfirmation std.Address) {
	caller := std.GetOrigCaller()
	if callerConfirmation != caller {
		panic("accesscontrol: caller confirmation does not match account")
	}
	roleData.Holder.Remove(caller.String())
	std.Emit("RoleRevoked", "roleName", roleData.Name, "account", caller.String(), "sender", caller.String())
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
	std.Emit("RoleAdminChanged", "roleName", roleData.Name, "previousAdminRole", previousAdminRole.String(), "newAdminRole", adminRole.String())
}
