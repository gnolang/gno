package disperse

import "errors"

var (
	errNotEnoughCoin            = errors.New("not enough coin sent in")
	errNumAddrValMismatch       = errors.New("number of addresses and values to send doesn't match")
	errInvalidAddress           = errors.New("invalid address")
	errNegativeCoinAmount       = errors.New("coin amount cannot be negative")
	errBalanceNotZero           = errors.New("balance needs to be equal to zero")
	ErrArgLenAndSentLenMismatch = errors.New("mismatch between coins sent and args called")
	ErrWrongAmount              = errors.New("wrong coin amount")
)