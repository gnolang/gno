package ownable2step

import "errors"

var (
	ErrNoPendingOwner      = errors.New("ownable2step: no pending owner")
	ErrUnauthorized        = errors.New("ownable2step: caller is not owner")
	ErrPendingUnauthorized = errors.New("ownable2step: caller is not pending owner")
	ErrInvalidAddress      = errors.New("ownable2step: new owner address is invalid")
)
