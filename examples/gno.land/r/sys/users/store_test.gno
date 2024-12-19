package users

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestIsOnWhitelist(t *testing.T) {
	t.Run("init caller registration", func(t *testing.T) {
		t.Parallel()
		uassert.True(t, IsOnWhitelist(whitelistedCallerAddr))
	})

	t.Run("not on whitelist", func(t *testing.T) {
		t.Parallel()
		uassert.False(t, IsOnWhitelist(aliceCodeRealm.Addr()))
	})
}

func TestWriteUser(t *testing.T) {
	t.Run("unauthorized write", func(t *testing.T) {
		t.Parallel()

		std.TestSetRealm(aliceCodeRealm)

		uassert.Error(t, SetUser(aliceUsername, alice), ErrUnauthorizedWrite.Error())
	})

	// todo tests:
	// 	test validate()
	// 	test already exists in ustore
	// 	test alreadyhasname
	//	test already taken as alias

	t.Run("already exists", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(std.NewCodeRealm(gnolandUsers))

		// write same user twice
		urequire.NoError(t, SetUser(aliceUsername, alice))
		uassert.Error(t, SetUser(aliceUsername, alice), ErrNameTaken.Error())
	})

	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(std.NewCodeRealm(gnolandUsers))

		uassert.Error(t, SetUser("test123", "invalidaddress"), ErrInvalidAddress.Error())
		uassert.Error(t, SetUser("test123", ""), ErrInvalidAddress.Error())
		uassert.Error(t, SetUser("test123",
			"000000000000000000000000000000000000000000000000"),
			ErrInvalidAddress.Error())
	})
}
