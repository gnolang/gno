// Package subacc (short for sub-accounts) implements an authenticator based on the caller and an user-provided slug.
package subacc

import (
	"path"
	"std"

	"gno.land/p/demo/auth"
	"gno.land/r/demo/authreg"
)

var source std.Realm

func init() {
	source = std.CurrentRealm()
	authreg.Register(Authenticate)
}

// AuthToken constructs an [auth.Token] based on the caller and a slug.
//
// Returned tokens have an entity ID namespaced with the caller address, preventing
// malicious callers from impersonating another address.
func AuthToken(slug string) auth.Token {
	caller := std.PrevRealm().Addr()
	return &token{accountKey: accountID(caller, slug)}
}

type token struct {
	accountKey string
}

func (a *token) Source() std.Realm {
	return source
}

var _ auth.Token = (*token)(nil)

// Authenticate implements [auth.AuthenticateFn]
func Authenticate(autok auth.Token) string {
	val, ok := autok.(*token)
	if !ok || val == nil {
		panic(auth.ErrInvalidToken)
	}

	return val.accountKey
}

var _ auth.AuthenticateFn = Authenticate

// EntityID returns the full entity ID for an address and a slug
func EntityID(addr std.Address, slug string) string {
	return path.Join("/", source.Addr().String(), accountID(addr, slug))
}

func accountID(creator std.Address, slug string) string {
	return auth.NamespacedEntityID(creator.String(), slug)
}
