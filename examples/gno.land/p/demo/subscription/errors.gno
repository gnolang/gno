package subscription

import "errors"

var (
	ErrPayment = errors.New("subscription: insufficient payment amount")
	ErrNoSub   = errors.New("subscription: no subscription found")
)
