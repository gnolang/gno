package users

import "errors"

var (
	ErrUnauthorizedWrite  = errors.New("r/sys/users: not authorized to write to store")
	ErrAlreadyExists      = errors.New("r/sys/users: username already exists")
	ErrAlreadyWhitelisted = errors.New("r/sys/users: already whitelisted")
	ErrNotWhitelisted     = errors.New("r/sys/users: does not exist in whitelist")
	ErrInvalidAddress     = errors.New("r/sys/users: invalid address")
	ErrEmptyUsername      = errors.New("r/sys/users: empty username provided")
	ErrUserIsNil          = errors.New("r/sys/users: user provided is nil")
)
