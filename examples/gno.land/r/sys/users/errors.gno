package users

import (
	"chain/runtime"
	"errors"

	"gno.land/p/nt/ufmt/v0"
)

const prefix = "r/sys/users: "

var (
	ErrAlreadyWhitelisted    = errors.New(prefix + "already whitelisted")
	ErrWhitelistRemoveFailed = errors.New(prefix + "failed to remove address from whitelist")

	ErrNameTaken          = errors.New(prefix + "name/Alias already taken")
	ErrCanonicalCollision = errors.New(prefix + "name collides with a confusable variant of an existing name")
	ErrInvalidAddress     = errors.New(prefix + "invalid address")

	ErrEmptyUsername   = errors.New(prefix + "empty username provided")
	ErrNameLikeAddress = errors.New(prefix + "username resembles a gno.land address")
	ErrInvalidUsername = errors.New(prefix + "username must match ^[a-z][a-z0-9]*([_-][a-z0-9]+)*$ (max 64 chars)")

	ErrAlreadyHasName = errors.New(prefix + "username for this address already registered - try creating an Alias")
	ErrDeletedUser    = errors.New(prefix + "cannot register a new username after deleting")

	ErrUserNotExistOrDeleted = errors.New(prefix + "this user does not exist or was deleted")
)

// ErrNotWhitelisted stores the current and previous realm identities as
// plain strings so the error is a pure data record with no live realm
// values held in its fields.
type ErrNotWhitelisted struct {
	Current  string // "CodeRealm{ <addr>, <pkgPath> }" or "UserRealm{ <addr> }" — not whitelisted
	Previous string // same shape — for context
}

func NewErrNotWhitelisted() ErrNotWhitelisted {
	return ErrNotWhitelisted{
		Current:  runtime.CurrentRealm().String(),
		Previous: runtime.PreviousRealm().String(),
	}
}

func (e ErrNotWhitelisted) Error() string {
	return ufmt.Sprintf("%scurrent realm/user does not exist in whitelist: %s (previous: %s)",
		prefix, e.Current, e.Previous)
}
