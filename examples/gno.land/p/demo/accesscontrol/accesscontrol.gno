package accesscontrol

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
)

const (
	RoleName     = "roleName"
	Sender       = "sender"
	Account      = "account"
	NewAdminRole = "newAdminRole"
)

// Role struct to store role information
type Role struct {
	Name    string
	Holders *avl.Tree // std.Address -> struct{}
	Ownable *ownable.Ownable
}

// Roles struct to store all Roles information
type Roles struct {
	AllRoles []*Role
	Ownable  *ownable.Ownable
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
		return nil, err
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

	if err := rs.Ownable.CallerIsOwner(); err != nil {
		return nil, err
	}

	for _, role := range rs.AllRoles {
		if role.Name == name {
			return nil, ErrRoleSameName
		}
	}

	role, err := NewRole(name, rs.Ownable.Owner())
	if err != nil {
		return nil, err
	}

	rs.AllRoles = append(rs.AllRoles, role)

	std.Emit(
		"RoleCreated",
		RoleName, name,
		Sender, rs.Ownable.Owner().String(),
	)

	return role, nil
}

// HasRole check if an account has a specific role
func (r *Role) HasRole(account std.Address) bool {
	return r.Holders.Has(account.String())
}

// FindRole searches for a role by its name
func (rs *Roles) FindRole(name string) (*Role, error) {
	for _, role := range rs.AllRoles {
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

	err = r.Ownable.CallerIsOwner()
	if err != nil {
		return err
	}

	r.Holders.Set(account.String(), struct{}{})

	std.Emit(
		"RoleGranted",
		RoleName, r.Name,
		Account, account.String(),
		Sender, std.PrevRealm().Addr().String(),
	)

	return nil
}

// RevokeRole revokes a role from an account
func (rs *Roles) RevokeRole(name string, account std.Address) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	err = r.Ownable.CallerIsOwner()
	if err != nil {
		return err
	}

	r.Holders.Remove(account.String())

	std.Emit(
		"RoleRevoked",
		RoleName, r.Name,
		Account, account.String(),
		Sender, std.PrevRealm().Addr().String(),
	)

	return nil
}

// RenounceRole allows an account to renounce a role it holds
func (rs *Roles) RenounceRole(name string) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	caller := std.PrevRealm().Addr()

	if !r.HasRole(caller) {
		return ErrAccountNotRole
	}

	r.Holders.Remove(caller.String())

	std.Emit(
		"RoleRenounced",
		RoleName, r.Name,
		Account, caller.String(),
		Sender, caller.String(),
	)

	return nil
}

// SetRoleAdmin transfers the ownership of the role to a new administrator
func (r *Role) SetRoleAdmin(admin std.Address) error {
	if err := r.Ownable.TransferOwnership(admin); err != nil {
		return err
	}

	std.Emit(
		"RoleSet",
		RoleName, r.Name,
		NewAdminRole, r.Ownable.Owner().String(),
	)

	return nil
}
