// Package accesscontrol provides a library for managing roles and
// permissions within Gno. It allows for the creation,
// assignment, and management of roles with specific admin privileges,
// ensuring that only authorized accounts can perform certain actions.
// This package utilizes an AVL tree data structure to efficiently track
// and manage role holders.
//
// Example Usage:
//
//  import "gno.land/p/demo/accesscontrol"
//
//  Create a new role with a specific admin.
//  adminRole := std.Address("admin-address")
//  role := accesscontrol.NewRoleData("ExampleRole", adminRole)
//
//  Check if an account has a specific role.
//  account := std.Address("user-address")
//  hasRole := role.HasRole(account)
//
//  Grant a role to a specific account.
//  role.GrantRole(account)
//
//  Revoke a role from a specific account.
//  role.RevokeRole(account)
//
//  Renounce a role with caller confirmation.
//  role.RenounceRole(std.GetOrigCaller())
//
//  Change the admin role for a specific role.
//  newAdmin := std.Address("new-admin-address")
//  role.SetRoleAdmin(newAdmin)

package accesscontrol // import "gno.land/p/demo/accesscontrol"

import (
	"std"
	"errors"
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
func (roleData *RoleData) AssertOrigCallerIsAdmin() {
	caller := std.GetOrigCaller()
	if roleData.AdminRole != caller {
		panic("accesscontrol: caller does not have the admin role")
	}

	return nil
}

// Method to create a new role within the realm
func (roleData *RoleData) CreateRole(name string, adminRole std.Address) *RoleData {
	roleData.AssertOrigCallerIsAdmin()
    std.Emit("RoleCreated", "roleName", name, "adminRole", adminRole.String(), "sender", std.GetOrigCaller().String())
    return NewRoleData(name, adminRole)
}

// Method to check if an account has a specific role
func (roleData *RoleData) HasRole(account std.Address) bool {
	return roleData.Holder.Has(account.String())
}

// Method to grant a role to an account
func (roleData *RoleData) GrantRole(account std.Address) {
	roleData.AssertOrigCallerIsAdmin()
	roleData.Holder.Set(account.String(), true)
	std.Emit("RoleGranted", "roleName", roleData.Name, "account", account.String(), "sender", std.GetOrigCaller().String())
}

// Method to revoke a role from an account
func (roleData *RoleData) RevokeRole(account std.Address) {
	roleData.AssertOrigCallerIsAdmin()
	roleData.Holder.Remove(account.String())
	std.Emit("RoleRevoked", "roleName", roleData.Name, "account", account.String(), "sender", std.GetOrigCaller().String())
}

// Method to renounce a role with caller confirmation
func (roleData *RoleData) RenounceRole(callerConfirmation std.Address) error {
	caller := std.GetOrigCaller()
	if callerConfirmation != caller {
		return errors.New("accesscontrol: caller confirmation does not match account")
	}
	roleData.Holder.Remove(caller.String())
	std.Emit("RoleRenounced", "roleName", roleData.Name, "account", caller.String(), "sender", caller.String())
	return nil
}

// Method to set the admin role for a specific role
func (roleData *RoleData) SetRoleAdmin(adminRole std.Address) {
	roleData.AssertOrigCallerIsAdmin()
	previousAdminRole := roleData.AdminRole
	roleData.AdminRole = adminRole
	std.Emit("RoleAdminChanged", "roleName", roleData.Name, "previousAdminRole", previousAdminRole.String(), "newAdminRole", adminRole.String())
}
