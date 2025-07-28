package users

import "errors"

var (
	ErrNonUserCall      = errors.New("registration must be done by a user")
	ErrPaused           = errors.New("registration is currently paused")
	ErrInvalidUsername  = errors.New("invalid username format")
	ErrInvalidPayment   = errors.New("invalid payment amount")
	ErrNameNotAvailable = errors.New("name is not available")
	ErrInvalidBidHash   = errors.New("invalid bid hash format")
)
