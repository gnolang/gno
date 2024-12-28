package auth_test

import (
	"std"
	"testing"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/urequire"
)

func TestToken(t *testing.T) {
	urequire.NotPanics(t, func() {
		authenticate(getToken())
	})

	urequire.PanicsWithMessage(t, auth.ErrInvalidToken.Error(), func() {
		authenticate((*token)(nil))
	})

	urequire.PanicsWithMessage(t, auth.ErrInvalidToken.Error(), func() {
		authenticate(getFakeToken())
	})

	urequire.PanicsWithMessage(t, auth.ErrInvalidToken.Error(), func() {
		authenticate(nil)
	})

	urequire.PanicsWithMessage(t, auth.ErrInvalidToken.Error(), func() {
		authenticate((*fakeToken)(nil))
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
	if val, ok := autok.(*token); !ok || val == nil {
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

func TestEntityID(t *testing.T) {
	cases := []struct {
		name         string
		provider     string
		subPath      string
		res          string
		panicMessage string
	}{
		{
			name:     "good",
			provider: "alice",
			subPath:  "savings",
			res:      "/alice/savings",
		},
		{
			name:         "mal_backtrack",
			provider:     "eve",
			subPath:      "../alice/savings",
			panicMessage: `invalid sub-path "../alice/savings"`,
		},
		{
			name:     "mal_backtrack_dumb",
			provider: "eve",
			subPath:  "/../alice/savings",
			res:      "/eve/alice/savings",
		},
		{
			name:         "mal_backtrack_hidden",
			provider:     "eve",
			subPath:      "todobien/very/deep/../../../../alice/savings",
			panicMessage: `invalid sub-path "../alice/savings"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := func() {
				res := entityID(tc.provider, tc.subPath)
				urequire.Equal(t, tc.res, res)
			}
			if tc.panicMessage != "" {
				urequire.PanicsWithMessage(t, tc.panicMessage, run)
			} else {
				urequire.NotPanics(t, run)
			}
		})
	}
}
