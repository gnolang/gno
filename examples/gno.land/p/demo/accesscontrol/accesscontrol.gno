package accesscontrol

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
)

var ErrUnauthorized = errors.New("unauthorized; caller is not admin")

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

// Method to check if the caller has the admin role and return an error
func (roleData *RoleData) CallerIsAdmin() error {
	caller := std.PrevRealm().Addr()
	if roleData.AdminRole != caller {
		return ErrUnauthorized
	}
	return nil
}

// Method to assert if the caller has the admin role, panics if not
func (roleData *RoleData) AssertCallerIsAdmin() {
	if err := roleData.CallerIsAdmin(); err != nil {
		panic(err)
	}
}

// Method to create a new role within the realm
func (roleData *RoleData) CreateRole(name string, adminRole std.Address) *RoleData {
	roleData.AssertCallerIsAdmin()
	std.Emit("RoleCreated", "roleName", name, "adminRole", adminRole.String(), "sender", std.PrevRealm().Addr().String())
	return NewRoleData(name, adminRole)
}

// Method to check if an account has a specific role
func (roleData *RoleData) HasRole(account std.Address) bool {
	return roleData.Holder.Has(account.String())
}

// Method to grant a role to an account
func (roleData *RoleData) GrantRole(account std.Address) {
	roleData.AssertCallerIsAdmin()
	roleData.Holder.Set(account.String(), true)
	std.Emit("RoleGranted", "roleName", roleData.Name, "account", account.String(), "sender", std.PrevRealm().Addr().String())
}

// Method to revoke a role from an account
func (roleData *RoleData) RevokeRole(account std.Address) {
	roleData.AssertCallerIsAdmin()
	roleData.Holder.Remove(account.String())
	std.Emit("RoleRevoked", "roleName", roleData.Name, "account", account.String(), "sender", std.PrevRealm().Addr().String())
}

// Method to renounce a role with caller confirmation
func (roleData *RoleData) RenounceRole(callerConfirmation std.Address) error {
	caller := std.PrevRealm().Addr()
	if callerConfirmation != caller {
		return errors.New("accesscontrol: caller confirmation does not match account")
	}
	roleData.Holder.Remove(caller.String())
	std.Emit("RoleRenounced", "roleName", roleData.Name, "account", caller.String(), "sender", caller.String())
	return nil
}

// Method to set the admin role for a specific role
func (roleData *RoleData) SetRoleAdmin(adminRole std.Address) {
	roleData.AssertCallerIsAdmin()
	previousAdminRole := roleData.AdminRole
	roleData.AdminRole = adminRole
	std.Emit("RoleAdminChanged", "roleName", roleData.Name, "previousAdminRole", previousAdminRole.String(), "newAdminRole", adminRole.String())
}
