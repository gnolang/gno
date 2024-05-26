package pausable

import "errors"

var (
	ErrUnauthorized   = errors.New("unauthorized; caller is not owner")
	ErrInvalidAddress = errors.New("new owner address is invalid")
)
