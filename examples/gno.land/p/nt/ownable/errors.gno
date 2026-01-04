package ownable

import "errors"

var (
	ErrUnauthorized   = errors.New("ownable: caller is not owner")
	ErrInvalidAddress = errors.New("ownable: new owner address is invalid")
)
