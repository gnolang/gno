package sessions

import (
	"std"
	"testing"

	"gno.land/p/demo/auth"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

// TODO: more extensive tests

func TestSessions(t *testing.T) {
	aliceAddr := testutils.TestAddress("alice")
	aliceRealm := std.NewUserRealm(aliceAddr)

	bobAddr := testutils.TestAddress("bob")
	bobRealm := std.NewUserRealm(bobAddr)

	// no session created
	{
		std.TestSetRealm(bobRealm)

		autok := AuthToken(aliceAddr)
		urequire.PanicsWithMessage(t, auth.ErrInvalidToken.Error(), func() {
			Authenticate(autok)
		})
	}

	// happy path
	{
		std.TestSetRealm(aliceRealm)

		Login(bobAddr, NoExpiry)

		std.TestSetRealm(bobRealm)

		autok := AuthToken(aliceAddr)
		entityID := Authenticate(autok)
		urequire.Equal(t, "/"+aliceAddr.String(), entityID)

		Logout(aliceAddr, bobAddr)

		autok = AuthToken(aliceAddr)
		urequire.PanicsWithMessage(t, auth.ErrInvalidToken.Error(), func() {
			Authenticate(autok)
		})
	}
}
