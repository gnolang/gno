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
	authreg.Register(authenticate)
}

func AuthToken(slug string) auth.Token {
	caller := std.PrevRealm().Addr()
	return &token{accountKey: accountKey(caller, slug)}
}

type token struct {
	accountKey string
}

func (a *token) Source() std.Realm {
	return source
}

var _ auth.Token = (*token)(nil)

func authenticate(autho auth.Token) string {
	// this check should ensure we created this object
	cauth, ok := autho.(*token)
	if !ok {
		panic(auth.ErrInvalidToken)
	}

	return cauth.accountKey
}

var _ auth.AuthenticateFn = authenticate

func accountKey(creator std.Address, slug string) string {
	return path.Join("/", creator.String(), slug)
}
