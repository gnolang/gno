// import "gno.land/p/demo/accesscontrol"
//
// Create a new role with a specific admin.
// adminAddr := std.Address("admin-address")
// role := accesscontrol.NewRole("ExampleRole", adminRole)
//
// Check if an account has a specific role.
// account := std.Address("user-address")
// hasRole := role.HasRole(account)
//
// Grant a role to a specific account.
// role.GrantRole(account)
//
// Revoke a role from a specific account.
// role.RevokeRole(account)
//
// Renounce a role with caller confirmation.
// role.RenounceRole(std.GetOrigCaller())
//
// Change the admin role for a specific role.
// newAdmin := std.Address("new-admin-address")
// role.SetRoleAdmin(newAdmin)
package accesscontrol

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
)

var ErrUnauthorized = errors.New("unauthorized; caller is not admin")

// Role struct to store role information
type Role struct {
	Name      string
	Holders   *avl.Tree // -> std.Address -> bool
	AdminRole std.Address
}

type Roles struct {
	AllRoles []*Role
	Admin    std.Address
}

// NewRole creates a new instance of Role
func NewRole(name string, adminRole std.Address) *Role {
	return &Role{
		Name:      name,
		Holders:   avl.NewTree(),
		AdminRole: adminRole,
	}
}

// Method to check if the caller has the admin role and return an error
func (r *Role) CallerIsAdmin() error {
	caller := std.PrevRealm().Addr()

	if r.AdminRole != caller {
		return ErrUnauthorized
	}

	return nil
}

// Method to assert if the caller has the admin role, panics if not
func (r *Role) AssertCallerIsAdmin() {
	if err := r.CallerIsAdmin(); err != nil {
		panic(err)
	}
}

// Method to create a new role within the realm
func (rs *Roles) CreateRole(name string, adminRole std.Address) *Role {
	if rs.Admin != std.PrevRealm().Addr() {
		panic("accesscontrol: caller does not have the global admin role")
	}

	role := NewRole(name, adminRole)
	rs.AllRoles = append(rs.AllRoles, role)

	std.Emit(
		"RoleCreated",
		"roleName", name,
		"adminRole", adminRole.String(),
		"sender", rs.Admin.String(),
	)

	return role
}

// Method to check if an account has a specific role
func (r *Role) HasRole(account std.Address) bool {
	return r.Holders.Has(account.String())
}

// Method to grant a role to an account
func (r *Role) GrantRole(account std.Address) {
	r.AssertCallerIsAdmin()
	r.Holders.Set(account.String(), true)

	std.Emit(
		"RoleGranted",
		"roleName", r.Name,
		"account", account.String(),
		"sender", std.PrevRealm().Addr().String(),
	)
}

// Method to revoke a role from an account
func (r *Role) RevokeRole(account std.Address) {
	r.AssertCallerIsAdmin()
	r.Holders.Remove(account.String())

	std.Emit(
		"RoleRevoked",
		"roleName", r.Name,
		"account", account.String(),
		"sender", std.PrevRealm().Addr().String(),
	)
}

// Method to renounce a role with caller confirmation
func (r *Role) RenounceRole(callerConfirmation std.Address) error {
	caller := std.PrevRealm().Addr()

	if callerConfirmation != caller {
		return errors.New("accesscontrol: caller confirmation does not match account")
	}

	r.Holders.Remove(caller.String())

	std.Emit(
		"RoleRenounced",
		"roleName", r.Name,
		"account", caller.String(),
		"sender", caller.String(),
	)

	return nil
}

// Method to set the admin role for a specific role
func (r *Role) SetRoleAdmin(adminRole std.Address) {
	r.AssertCallerIsAdmin()

	previousAdminRole := r.AdminRole
	r.AdminRole = adminRole

	std.Emit(
		"RoleSet",
		"roleName", r.Name,
		"previousAdminRole", previousAdminRole.String(),
		"newAdminRole", adminRole.String(),
	)
}
