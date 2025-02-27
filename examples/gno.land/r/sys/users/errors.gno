package users

import "errors"

const prefix = "r/sys/users: "

var (
	ErrNotWhitelisted     = errors.New(prefix + "does not exist in whitelist")
	ErrAlreadyWhitelisted = errors.New(prefix + "already whitelisted")

	ErrNameTaken      = errors.New(prefix + "name/Alias already taken")
	ErrInvalidAddress = errors.New(prefix + "invalid address")

	ErrEmptyUsername   = errors.New(prefix + "empty username provided")
	ErrNameLikeAddress = errors.New(prefix + "username resembles a gno.land address")

	ErrAlreadyHasName = errors.New(prefix + "username for this address already registered - try creating an Alias")
	ErrDeletedUser    = errors.New(prefix + "cannot register a new username after deleting")

	ErrUserNotExistOrDeleted = errors.New(prefix + "this user does not exist or was deleted")
)
