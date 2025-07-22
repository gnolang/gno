package staker

import (
	"errors"
)

var (
	ErrInvalidAmount = errors.New("amount must be positive")
	ErrInvalidDelegatee = errors.New("invalid delegatee address")
	ErrInsufficientAllowance = errors.New("insufficient allowance")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrCooldownNotFinished = errors.New("cooldown not finished")
) 
