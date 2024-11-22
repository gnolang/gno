package users

import "errors"

var (
	ErrUnauthorizedWrite  = errors.New("r/sys/users: not authorized to write to store")
	ErrAlreadyExists      = errors.New("r/sys/users: username already exists")
	ErrAlreadyWhitelisted = errors.New("r/sys/users: already whitelisted")
	ErrDoesNotExist       = errors.New("r/sys/users: does not exist in whitelist")
	ErrInvalidAddress     = errors.New("r/sys/users: invalid address")
)
