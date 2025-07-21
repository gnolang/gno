package vls

import "std"

var (
	ErrOnlyGovernanceCanExecute = std.Errorf("only governance contract can execute")
	ErrOnlyGovernanceCanUpdate  = std.Errorf("only current governance contract can update governance")
	ErrInsufficientBalance      = std.Errorf("insufficient balance")
	ErrInsufficientAllowance    = std.Errorf("insufficient allowance")
	ErrInvalidAmount            = std.Errorf("invalid amount")
	ErrInvalidAddress           = std.Errorf("invalid address")
	ErrMintOverflow             = std.Errorf("mint overflow")
	ErrCannotTransferToSelf     = std.Errorf("cannot transfer to self")
	ErrAirdropTreeEmpty         = std.Errorf("airdrop tree is empty")
) 
