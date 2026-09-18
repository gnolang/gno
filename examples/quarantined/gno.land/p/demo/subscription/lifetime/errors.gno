package lifetime

import "errors"

var (
	ErrNoSub             = errors.New("lifetime subscription: no active subscription found")
	ErrAlreadySub        = errors.New("lifetime subscription: this address already has an active lifetime subscription")
	ErrNotAuthorized     = errors.New("lifetime subscription: action not authorized")
	ErrInvalidPayee      = errors.New("lifetime subscription: payee address is invalid")
	ErrInvalidSubscriber = errors.New("lifetime subscription: subscriber address is invalid")
)
