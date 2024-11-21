package users

import "errors"

var (
	ErrUnauthorized  = errors.New("sys/users: not authorized")
	ErrAlreadyExists = errors.New("users: username already exists")
)
