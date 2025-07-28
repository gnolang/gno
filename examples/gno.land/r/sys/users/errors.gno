package users

import (
	"errors"
	"std"

	"gno.land/p/demo/ufmt"
)

const prefix = "r/sys/users: "

var (
	ErrAlreadyWhitelisted = errors.New(prefix + "already whitelisted")

	ErrNameTaken      = errors.New(prefix + "name/Alias already taken")
	ErrInvalidAddress = errors.New(prefix + "invalid address")

	ErrEmptyUsername   = errors.New(prefix + "empty username provided")
	ErrNameLikeAddress = errors.New(prefix + "username resembles a gno.land address")
	ErrInvalidUsername = errors.New(prefix + "username must match ^[a-zA-Z0-9_]{1,64}$")

	ErrAlreadyHasName = errors.New(prefix + "username for this address already registered - try creating an Alias")
	ErrDeletedUser    = errors.New(prefix + "cannot register a new username after deleting")

	ErrUserNotExistOrDeleted = errors.New(prefix + "this user does not exist or was deleted")

	// Anti-squatting related errors
	ErrNameRequiresAuction     = errors.New(prefix + "name requires auction-based registration")
	ErrAntiSquattingDisabled   = errors.New(prefix + "anti-squatting system is disabled")
	ErrSystemNotConfigured     = errors.New(prefix + "anti-squatting system not properly configured")
	ErrNameNotAvailable        = errors.New(prefix + "name is not available for registration")
	ErrNameNotFound            = errors.New(prefix + "name not found")
	ErrInsufficientDeposit     = errors.New(prefix + "insufficient deposit for auction bid")
	ErrInsufficientDisputeFee  = errors.New(prefix + "insufficient fee for dispute filing")
)

type ErrNotWhitelisted struct {
	Current  std.Realm // not whitelisted
	Previous std.Realm // for context
}

func NewErrNotWhitelisted() ErrNotWhitelisted {
	return ErrNotWhitelisted{
		Current:  std.CurrentRealm(),
		Previous: std.PreviousRealm(),
	}
}

func (e ErrNotWhitelisted) Error() string {
	return ufmt.Sprintf("%scurrent realm/user does not exist in whitelist: %v (previous: %v)",
		prefix, e.Current, e.Previous)
}
