package auth

import (
	"errors"
	"std"
)

// Token represents an authentication token
type Token interface {
	// Source is used to identify the provenance of the token.
	// It is intented to be used to find the corresponding [AuthenticateFn], for example in registries.
	// It can be spoofed and should not be trusted.
	Source() std.Realm
}

// AuthenticateFn validates a token and returns the ID of the authenticated entity.
// It panics with [ErrInvalidToken] if the token is invalid.
//
// XXX: Should we add some kind of `Token.EntityID` method instead of returning the ID here?
type AuthenticateFn = func(autok Token) string

var (
	ErrInvalidToken = errors.New("invalid token")
)
