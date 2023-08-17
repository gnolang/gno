package gvotes

import "errors"

var (
	ErrInsufficientVotes    = errors.New("insufficient votes")
	ErrInvalidAddress       = errors.New("invalid address")
	ErrCannotTransferToSelf = errors.New("cannot send transfer to self")
)
