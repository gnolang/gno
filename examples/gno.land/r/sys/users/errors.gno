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

// SECURITY XXX: ErrNotWhitelisted holds runtime.Realm values as struct
// fields. The error is read-only and only used for formatting an error
// message (via .Error() below), so there's no active capability leak
// here today. BUT — storing realm values in struct fields is the same
// anti-pattern as storing `cur` itself. Once this pattern exists,
// future code might be tempted to actually USE the Realm methods on
// the stored value (Address(), PkgPath() — currently fine; or in the
// future, anything that walks back to identity). The Realm value's
// IsCurrent() always returns false for a captured value, which is the
// only thing preventing it from being a live capability today.
// Fix: store the PkgPath and Address as plain strings, not the full
// Realm struct. Same display behavior, no anti-pattern footprint.
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
