package auth_test

import (
	"std"
	"testing"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/urequire"
)

func TestValidToken(t *testing.T) {
	urequire.NotPanics(t, func() {
		authenticate(getToken())
	})
}

func TestFakeToken(t *testing.T) {
	urequire.PanicsWithMessage(t, auth.ErrInvalidToken.Error(), func() {
		authenticate(getFakeToken())
	})
}

var testRealm = std.NewCodeRealm("gno.land/r/demo/absacc")

type token struct {
}

func (t *token) Source() std.Realm {
	return testRealm
}

var _ auth.Token = (*token)(nil)

func getToken() auth.Token {
	return &token{}
}

func authenticate(autok auth.Token) string {
	// the next line is the core of the auth pattern, this ensures we created this token
	_, ok := autok.(*token)
	if !ok {
		panic(auth.ErrInvalidToken)
	}
	return "alice"
}

var _ auth.AuthenticateFn = authenticate

type fakeToken struct {
}

func (t *fakeToken) Source() std.Realm {
	return testRealm
}

var _ auth.Token = (*fakeToken)(nil)

func getFakeToken() auth.Token {
	return &fakeToken{}
}
