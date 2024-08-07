package disperse

import "errors"

var (
	ErrNotEnoughCoin            = errors.New("disperse: not enough coin sent in")
	ErrNumAddrValMismatch       = errors.New("disperse: number of addresses and values to send doesn't match")
	ErrInvalidAddress           = errors.New("disperse: invalid address")
	ErrNegativeCoinAmount       = errors.New("disperse: coin amount cannot be negative")
	ErrBalanceNotEqToAmountSent = errors.New("disperse: balance needs to be equal to amount sent")
	ErrArgLenAndSentLenMismatch = errors.New("disperse: mismatch between coins sent and args called")
	ErrWrongAmount              = errors.New("disperse: wrong coin amount")
)
