package auth

import (
	"errors"
	"std"
)

type Token interface {
	Source() std.Realm // this can be spoofed and is only useful for registries
}

type AuthenticateFn = func(auth Token) string

var (
	ErrInvalidToken = errors.New("invalid token")
)
