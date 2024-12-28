package authreg

import (
	"std"
	"testing"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/urequire"
)

func TestAuthreg(t *testing.T) {
	autok := &testAuthToken{source: std.NewCodeRealm("gno.land/r/testing/tester")}
	unknownAutok := &testAuthToken{source: std.NewCodeRealm("gno.land/r/testing/unknown")}
	called := false
	mockAuthFn := func(autok auth.Token) string {
		called = true
		return "alice"
	}

	std.TestSetRealm(autok.Source())
	Register(mockAuthFn)

	val := Authenticate(autok)
	urequire.True(t, called)
	expected := ufmt.Sprintf("/%s/%s", autok.Source().Addr().String(), "alice")
	urequire.Equal(t, expected, val)

	called = false
	urequire.PanicsWithMessage(t, "unknown auth provider", func() {
		Authenticate(unknownAutok)
	})
	urequire.False(t, called)
}

type testAuthToken struct {
	source std.Realm
}

func (t *testAuthToken) Source() std.Realm {
	return t.source
}

var _ auth.Token = (*testAuthToken)(nil)
