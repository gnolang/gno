package users

import (
	"errors"
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

// ErrNotWhitelisted stores the failing caller's realm identity as a
// plain string so the error is a pure data record (no live realm values
// in its fields).
type ErrNotWhitelisted struct {
	Caller string // "CodeRealm{ <addr>, <pkgPath> }" or "UserRealm{ <addr> }" — failed the whitelist check
}

// NewErrNotWhitelisted constructs the error with the caller's realm
// identity captured as a string at construction time. The _ int
// discriminator keeps this non-crossing (a non-crossing function can't
// take a `realm`-named-`cur` first param, so we use the standard
// _ int, rlm realm shape).
func NewErrNotWhitelisted(_ int, caller realm) ErrNotWhitelisted {
	return ErrNotWhitelisted{
		Caller: caller.String(),
	}
}

func (e ErrNotWhitelisted) Error() string {
	return prefix + "caller realm/user does not exist in whitelist: " + e.Caller
}
