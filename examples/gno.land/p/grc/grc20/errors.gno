package grc20

import "errors"

var (
	ErrInsufficientBalance  = errors.New("insufficient allowance")
	ErrInvalidAddress       = errors.New("invalid address")
	ErrCannotTransferToSelf = errors.New("cannot send transfer to self")
)
