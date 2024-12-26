package auth

import (
	"errors"
	"std"
)

type Token interface {
	Source() std.Realm // this can be spoofed and is only useful for registries
}

// AuthenticateFn validates a token and returns the ID of the authenticated entity
type AuthenticateFn = func(auth Token) string

var (
	ErrInvalidToken = errors.New("invalid token")
)
