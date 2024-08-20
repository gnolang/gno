package lifetime

import "errors"

var (
	ErrNoSub         = errors.New("subscription lifetime: no subscription found")
	ErrAmt           = errors.New("subscription lifetime: the amount does not correspond")
	ErrAlreadySub    = errors.New("subscription lifetime: this address already has a lifetime subscription")
	ErrNotAuthorized = errors.New("subscription lifetime: action not authorized")
)
