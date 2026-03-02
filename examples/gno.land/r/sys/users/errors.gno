package users

import (
	"chain/runtime"
	"errors"

	"gno.land/p/nt/ufmt"
)

const prefix = "r/sys/users: "

var (
	ErrAlreadyWhitelisted    = errors.New(prefix + "already whitelisted")
	ErrWhitelistRemoveFailed = errors.New(prefix + "failed to remove address from whitelist")

	ErrNameTaken      = errors.New(prefix + "name/Alias already taken")
	ErrInvalidAddress = errors.New(prefix + "invalid address")

	ErrEmptyUsername   = errors.New(prefix + "empty username provided")
	ErrNameLikeAddress = errors.New(prefix + "username resembles a gno.land address")
	ErrInvalidUsername = errors.New(prefix + "username must match ^[a-zA-Z0-9_]{1,64}$")

	ErrAlreadyHasName = errors.New(prefix + "username for this address already registered - try creating an Alias")
	ErrDeletedUser    = errors.New(prefix + "cannot register a new username after deleting")

	ErrUserNotExistOrDeleted = errors.New(prefix + "this user does not exist or was deleted")
)

type ErrNotWhitelisted struct {
	Current  runtime.Realm // not whitelisted
	Previous runtime.Realm // for context
}

func NewErrNotWhitelisted() ErrNotWhitelisted {
	return ErrNotWhitelisted{
		Current:  runtime.CurrentRealm(),
		Previous: runtime.PreviousRealm(),
	}
}

func (e ErrNotWhitelisted) Error() string {
	return ufmt.Sprintf("%scurrent realm/user does not exist in whitelist: %v (previous: %v)",
		prefix, e.Current, e.Previous)
}
