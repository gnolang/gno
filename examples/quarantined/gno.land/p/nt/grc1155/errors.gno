package grc1155

import "errors"

var (
	ErrInvalidAddress                         = errors.New("invalid address")
	ErrMismatchLength                         = errors.New("accounts and ids length mismatch")
	ErrCannotTransferToSelf                   = errors.New("cannot send transfer to self")
	ErrTransferToRejectedOrNonGRC1155Receiver = errors.New("transfer to rejected or non GRC1155Receiver implementer")
	ErrCallerIsNotOwnerOrApproved             = errors.New("caller is not token owner or approved")
	ErrInsufficientBalance                    = errors.New("insufficient balance for transfer")
	ErrBurnAmountExceedsBalance               = errors.New("burn amount exceeds balance")
	ErrInvalidAmount                          = errors.New("invalid amount")
)
