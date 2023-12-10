package disperse

import "errors"

var (
	errNotEnoughGnot      = errors.New("not enough gnot sent in")
	errNumAddrValMismatch = errors.New("number of addresses and values to send doesn't match")
	errInvalidAddress     = errors.New("invalid address")
	errNegativeCoinAmount = errors.New("coin amount cannot be negative")
)
