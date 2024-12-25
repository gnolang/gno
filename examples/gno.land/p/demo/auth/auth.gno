package auth

import (
	"errors"
	"std"
)

type Token interface {
	Source() std.Realm
}

type AuthenticateFn = func(auth Token) string

var (
	ErrInvalidToken = errors.New("invalid token")
)
