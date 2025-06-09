package recurring

import "errors"

var (
	ErrNoSub         = errors.New("recurring subscription: no active subscription found")
	ErrSubExpired    = errors.New("recurring subscription: your subscription has expired")
	ErrAmt           = errors.New("recurring subscription: payment amount does not match the required subscription amount")
	ErrAlreadySub    = errors.New("recurring subscription: this address already has an active subscription")
	ErrNotAuthorized = errors.New("recurring subscription: action not authorized")
)
