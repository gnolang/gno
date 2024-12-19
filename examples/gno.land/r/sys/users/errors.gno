package users

import "errors"

const prefix = "r/sys/users: "

var (
	ErrUnauthorizedWrite  = errors.New(prefix + "not authorized to write to store")
	ErrNotWhitelisted     = errors.New(prefix + "does not exist in whitelist")
	ErrAlreadyWhitelisted = errors.New(prefix + "already whitelisted")

	ErrAlreadyExists  = errors.New(prefix + "name/alias already taken")
	ErrInvalidAddress = errors.New(prefix + "invalid address")

	ErrEmptyUsername       = errors.New(prefix + "empty username provided")
	ErrUsernameTooLong     = errors.New(prefix + "username is too long")
	ErrUsernameLikeAddress = errors.New(prefix + "username resembles a gno.land address")

	ErrAlreadyHasName  = errors.New(prefix + "username for this address already registered")
	ErrDeletedUser     = errors.New(prefix + "cannot register a new username after deleting")
	ErrAlreadyDeleted  = errors.New(prefix + "cannot delete same user twice")
	ErrAliasBeforeName = errors.New(prefix + "cannot register alias before a username")

	ErrUserNotExist = errors.New(prefix + "this user has not been registered before")
)
