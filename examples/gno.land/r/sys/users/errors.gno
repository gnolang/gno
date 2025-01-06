package users

import "errors"

const prefix = "r/sys/users: "

var (
	ErrNotWhitelisted     = errors.New(prefix + "does not exist in whitelist")
	ErrAlreadyWhitelisted = errors.New(prefix + "already whitelisted")

	ErrNameTaken      = errors.New(prefix + "name/Alias already taken")
	ErrInvalidAddress = errors.New(prefix + "invalid address")

	ErrEmptyUsername   = errors.New(prefix + "empty Username provided")
	ErrNameTooLong     = errors.New(prefix + "Username is too long")
	ErrNameLikeAddress = errors.New(prefix + "Username resembles a gno.land address")

	ErrAlreadyHasName  = errors.New(prefix + "Username for this address already registered - try creating an Alias")
	ErrDeletedUser     = errors.New(prefix + "cannot register a new Username after deleting")
	ErrAlreadyDeleted  = errors.New(prefix + "cannot delete same user twice")
	ErrAliasBeforeName = errors.New(prefix + "cannot register Alias before a Username")

	ErrUserNotExist = errors.New(prefix + "this user has not been registered before")
)
