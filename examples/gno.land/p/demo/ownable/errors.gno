package ownable

import "errors"

var (
	ErrUnauthorized   = errors.New("unauthorized; address is not owner")
	ErrInvalidAddress = errors.New("new owner address is invalid")
)
