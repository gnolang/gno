package accesscontrol

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
)

const (
	RoleCreatedEvent   = "RoleCreated"
	RoleGrantedEvent   = "RoleGranted"
	RoleRevokedEvent   = "RoleRevoked"
	RoleRenouncedEvent = "RoleRenounced"
	RoleSetEvent       = "RoleSet"
)

// Role struct to store role information
type Role struct {
	Name    string
	Holders *avl.Tree // std.Address -> struct{}
	Ownable *ownable.Ownable
}

// Roles struct to store all Roles information
type Roles struct {
	Roles       []*Role
	UserToRoles avl.Tree // std.Address -> []*Role
	Ownable     *ownable.Ownable
}

func validRoleName(name string) error {
	if len(name) > 30 || name == "" {
		return ErrNameRole
	}
	return nil
}

// NewRole creates a new instance of Role
func NewRole(name string, admin std.Address) (*Role, error) {
	if err := validRoleName(name); err != nil {
		return nil, ErrNameRole
	}

	return &Role{
		Name:    name,
		Holders: avl.NewTree(),
		Ownable: ownable.NewWithAddress(admin),
	}, nil
}

// CreateRole create a new role within the realm
func (rs *Roles) CreateRole(name string) (*Role, error) {
	if err := validRoleName(name); err != nil {
		return nil, ErrNameRole
	}

	if !rs.Ownable.OwnedByCurrent() {
		return nil, ErrNotOwner
	}

	for _, role := range rs.Roles {
		if role.Name == name {
			return nil, ErrRoleSameName
		}
	}

	role, err := NewRole(name, rs.Ownable.Owner())
	if err != nil {
		return nil, err
	}

	rs.Roles = append(rs.Roles, role)

	std.Emit(
		RoleCreatedEvent,
		"roleName", name,
		"sender", rs.Ownable.Owner().String(),
	)

	return role, nil
}

// HasAccount check if an account has a specific role
func (r *Role) HasAccount(account std.Address) bool {
	return r.Holders.Has(account.String())
}

// FindRole searches for a role by its name
func (rs *Roles) FindRole(name string) (*Role, error) {
	for _, role := range rs.Roles {
		if role.Name == name {
			return role, nil
		}
	}

	return nil, ErrRoleNotFound
}

// GrantRole grants a role to an account
func (rs *Roles) GrantRole(name string, account std.Address) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	if !r.Ownable.OwnedByCurrent() {
		return ErrNotOwner
	}

	r.Holders.Set(account.String(), struct{}{})

	// Add in UserToRoles
	roles, found := rs.UserToRoles.Get(account.String())
	if !found {
		roles = []*Role{}
	}
	roles = append(roles.([]*Role), r)
	rs.UserToRoles.Set(account.String(), roles)

	std.Emit(
		RoleGrantedEvent,
		"roleName", r.Name,
		"account", account.String(),
		"sender", std.CurrentRealm().Address().String(),
	)

	return nil
}

// RevokeRole revokes a role from an account
func (rs *Roles) RevokeRole(name string, account std.Address) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	if !r.Ownable.OwnedByCurrent() {
		return ErrNotOwner
	}

	r.Holders.Remove(account.String())

	// Remove in UserToRoles
	roles, found := rs.UserToRoles.Get(account.String())
	if found {
		updatedRoles := []*Role{}
		for _, role := range roles.([]*Role) {
			if role != r {
				updatedRoles = append(updatedRoles, role)
			}
		}
		rs.UserToRoles.Set(account.String(), updatedRoles)
	}

	std.Emit(
		RoleRevokedEvent,
		"roleName", r.Name,
		"account", account.String(),
		"sender", std.CurrentRealm().Address().String(),
	)

	return nil
}

// RenounceRole allows an account to renounce a role it holds
func (rs *Roles) RenounceRole(name string) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	caller := std.OriginCaller()

	if !r.HasAccount(caller) {
		return ErrAccountNotRole
	}

	r.Holders.Remove(caller.String())

	std.Emit(
		RoleRenouncedEvent,
		"roleName", r.Name,
		"account", caller.String(),
		"sender", caller.String(),
	)

	return nil
}

// SetRoleAdmin transfers the ownership of the role to a new administrator
func (r *Role) SetRoleAdmin(admin std.Address) error {
	if err := r.Ownable.TransferOwnership(admin); err != nil {
		return err
	}

	std.Emit(
		RoleSetEvent,
		"roleName", r.Name,
		"newAdminRole", r.Ownable.Owner().String(),
	)

	return nil
}
