package lifetime

import "errors"

var (
	ErrNoSub         = errors.New("lifetime subscription: no active subscription found")
	ErrAmt           = errors.New("lifetime subscription: payment amount does not match the required subscription amount")
	ErrAlreadySub    = errors.New("lifetime subscription: this address already has an active lifetime subscription")
	ErrNotAuthorized = errors.New("lifetime subscription: action not authorized")
	ErrNonUserCall   = errors.New("lifetime subscription: only user-call (maketx call) accepted")
)
