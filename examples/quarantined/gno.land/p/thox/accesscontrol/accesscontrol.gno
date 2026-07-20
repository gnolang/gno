package accesscontrol

import (
	"chain"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
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
	Holders *avl.Tree // address -> struct{}
	Ownable *ownable.Ownable
}

// Roles struct to store all Roles information
type Roles struct {
	Roles       []*Role
	UserToRoles avl.Tree // address -> []*Role
	Ownable     *ownable.Ownable
}

func validRoleName(name string) error {
	if len(name) > 30 || name == "" {
		return ErrNameRole
	}
	return nil
}

// NewRole creates a new instance of Role
func NewRole(name string, admin address) (*Role, error) {
	if err := validRoleName(name); err != nil {
		return nil, ErrNameRole
	}

	return &Role{
		Name:    name,
		Holders: avl.NewTree(),
		Ownable: ownable.NewWithAddress(admin),
	}, nil
}

// CreateRole create a new role within the realm. rlm must be the
// caller's own captured cur; rlm.Previous().Address() must equal the
// owner of the Roles container.
func (rs *Roles) CreateRole(_ int, rlm realm, name string) (*Role, error) {
	if err := validRoleName(name); err != nil {
		return nil, ErrNameRole
	}

	if !rlm.IsCurrent() {
		return nil, ErrNotOwner
	}
	if !rs.Ownable.OwnedBy(rlm.Previous().Address()) {
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

	chain.Emit(
		RoleCreatedEvent,
		"roleName", name,
		"sender", rs.Ownable.Owner().String(),
	)

	return role, nil
}

// HasAccount check if an account has a specific role
func (r *Role) HasAccount(account address) bool {
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

// GrantRole grants a role to an account. rlm must be the caller's own
// captured cur; rlm.Previous().Address() must equal the role's
// admin/owner.
func (rs *Roles) GrantRole(_ int, rlm realm, name string, account address) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	if !rlm.IsCurrent() {
		return ErrNotOwner
	}
	caller := rlm.Previous().Address()
	if !r.Ownable.OwnedBy(caller) {
		return ErrNotOwner
	}

	r.Holders.Set(account.String(), struct{}{})

	// Add in UserToRoles
	roles := rs.UserToRoles.Get(account.String())
	if roles == nil {
		roles = []*Role{}
	}
	roles = append(roles.([]*Role), r)
	rs.UserToRoles.Set(account.String(), roles)

	chain.Emit(
		RoleGrantedEvent,
		"roleName", r.Name,
		"account", account.String(),
		"sender", caller.String(),
	)

	return nil
}

// RevokeRole revokes a role from an account. rlm must be the caller's
// own captured cur; rlm.Previous().Address() must equal the role's
// admin/owner.
func (rs *Roles) RevokeRole(_ int, rlm realm, name string, account address) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	if !rlm.IsCurrent() {
		return ErrNotOwner
	}
	caller := rlm.Previous().Address()
	if !r.Ownable.OwnedBy(caller) {
		return ErrNotOwner
	}

	r.Holders.Remove(account.String())

	// Remove in UserToRoles
	roles := rs.UserToRoles.Get(account.String())
	if roles != nil {
		updatedRoles := []*Role{}
		for _, role := range roles.([]*Role) {
			if role != r {
				updatedRoles = append(updatedRoles, role)
			}
		}
		rs.UserToRoles.Set(account.String(), updatedRoles)
	}

	chain.Emit(
		RoleRevokedEvent,
		"roleName", r.Name,
		"account", account.String(),
		"sender", caller.String(),
	)

	return nil
}

// RenounceRole allows an account to renounce a role it holds. rlm must
// be the caller's own captured cur; the renouncer is
// rlm.Previous().Address(). Closing the forgery here matters: without
// it, any external realm could force a victim to renounce a role by
// passing the victim's address as caller.
func (rs *Roles) RenounceRole(_ int, rlm realm, name string) error {
	r, err := rs.FindRole(name)
	if err != nil {
		return ErrRoleNotFound
	}

	if !rlm.IsCurrent() {
		return ErrAccountNotRole
	}
	caller := rlm.Previous().Address()
	if !r.HasAccount(caller) {
		return ErrAccountNotRole
	}

	r.Holders.Remove(caller.String())

	chain.Emit(
		RoleRenouncedEvent,
		"roleName", r.Name,
		"account", caller.String(),
		"sender", caller.String(),
	)

	return nil
}

// SetRoleAdmin transfers the ownership of the role to a new
// administrator. rlm must be the caller's own captured cur and must be
// the current admin (delegated to ownable.TransferOwnership which
// re-checks via rlm.Previous().Address()).
func (r *Role) SetRoleAdmin(_ int, rlm realm, admin address) error {
	if err := r.Ownable.TransferOwnership(0, rlm, admin); err != nil {
		return err
	}

	chain.Emit(
		RoleSetEvent,
		"roleName", r.Name,
		"newAdminRole", r.Ownable.Owner().String(),
	)

	return nil
}
