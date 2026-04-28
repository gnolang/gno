package lifetime

import "errors"

var (
	ErrNoSub         = errors.New("lifetime subscription: no active subscription found")
	ErrAlreadySub    = errors.New("lifetime subscription: this address already has an active lifetime subscription")
	ErrNotAuthorized = errors.New("lifetime subscription: action not authorized")
)
