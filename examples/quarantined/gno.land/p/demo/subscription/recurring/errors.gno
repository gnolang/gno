package recurring

import "errors"

var (
	ErrNoSub             = errors.New("recurring subscription: no active subscription found")
	ErrSubExpired        = errors.New("recurring subscription: your subscription has expired")
	ErrAlreadySub        = errors.New("recurring subscription: this address already has an active subscription")
	ErrNotAuthorized     = errors.New("recurring subscription: action not authorized")
	ErrInvalidPayee      = errors.New("recurring subscription: payee address is invalid")
	ErrInvalidSubscriber = errors.New("recurring subscription: subscriber address is invalid")
)
