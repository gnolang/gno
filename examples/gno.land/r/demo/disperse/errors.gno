package disperse

import "errors"

var (
	ErrNotEnoughCoin            = errors.New("not enough coin sent in")
	ErrNumAddrValMismatch       = errors.New("number of addresses and values to send doesn't match")
	ErrInvalidAddress           = errors.New("invalid address")
	ErrNegativeCoinAmount       = errors.New("coin amount cannot be negative")
	ErrBalanceNotEqToAmountSent = errors.New("balance needs to be equal to amount sent")
	ErrArgLenAndSentLenMismatch = errors.New("mismatch between coins sent and args called")
	ErrWrongAmount              = errors.New("wrong coin amount")
)
