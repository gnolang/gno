package vls

import (
	"errors"
)

var (
	ErrOnlyGovernanceCanExecute = errors.New("only governance contract can execute")
	ErrOnlyGovernanceCanUpdate  = errors.New("only current governance contract can update governance")
	ErrInsufficientBalance      = errors.New("insufficient balance")
	ErrInsufficientAllowance    = errors.New("insufficient allowance")
	ErrInvalidAmount            = errors.New("invalid amount")
	ErrInvalidAddress           = errors.New("invalid address")
	ErrMintOverflow             = errors.New("mint overflow")
)
