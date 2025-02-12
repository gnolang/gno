package referral

import (
    "errors"
)

var (
    ErrInvalidAddress  = errors.New("invalid address format")
    ErrZeroAddress     = errors.New("zero address is not allowed")
    ErrSelfReferral    = errors.New("self referral is not allowed")
    ErrUnauthorized    = errors.New("unauthorized caller")
    ErrInvalidCaller   = errors.New("invalid caller")
    ErrCyclicReference = errors.New("cyclic reference is not allowed")
    ErrTooManyRequests = errors.New("too many requests: operations allowed once per 24 hours for each address")
    ErrNotFound        = errors.New("referral not found")
)
