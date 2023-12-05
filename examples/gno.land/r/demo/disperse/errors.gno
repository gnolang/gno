package disperse

import "errors"

var (
	ErrNotEnoughGnot      = errors.New("not enough gnot sent in")
	ErrNumAddrValMismatch = errors.New("number of addresses and values to send doesn't match")
	ErrInvalidAddress     = errors.New("invalid address")
)
