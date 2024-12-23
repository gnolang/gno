package accesscontrol

import "errors"

var (
	ErrNotMatchAccount = errors.New("accesscontrol: caller confirmation does not match account")
	ErrRoleSameName    = errors.New("accesscontrol: role already exists with the same name")
	ErrRoleNotFound    = errors.New("accesscontrol: role not found")
	ErrAccountNotRole  = errors.New("accesscontrol: this account does not have the role")
	ErrNameRole        = errors.New("accesscontrol: role name cannot be empty or exceed 30 characters")
)
