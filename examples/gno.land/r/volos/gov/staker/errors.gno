package staker

import (
	"errors"
)

var (
	ErrInvalidAmount          = errors.New("amount must be positive")
	ErrInvalidDelegatee       = errors.New("invalid delegatee address")
	ErrInsufficientAllowance  = errors.New("insufficient allowance")
	ErrInsufficientBalance    = errors.New("insufficient balance")
	ErrNoReadyUnstake         = errors.New("no ready unstake")
	ErrInsufficientDelegation = errors.New("insufficient delegation")
)
