package disperse

import "errors"

var (
	errNotEnoughCoin      = errors.New("not enough coin sent in")
	errNumAddrValMismatch = errors.New("number of addresses and values to send doesn't match")
	errInvalidAddress     = errors.New("invalid address")
	errNegativeCoinAmount = errors.New("coin amount cannot be negative")
	errMultipleCoinsSent  = errors.New("cannot send multiple tokens")
	errBalanceNotZero     = errors.New("balance needs to be equal to zero")
)
