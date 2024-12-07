package users

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	susers "gno.land/p/sys/users"
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

		u := susers.User{}
		std.TestSetRealm(aliceCodeRealm)

		uassert.Error(t, WriteUser(u), ErrUnauthorizedWrite.Error())
	})

	t.Run("already exists", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(std.NewCodeRealm(gnolandUsers))

		u := susers.User{
			Name:    aliceUsername,
			Address: alice,
		}

		// write same user twice
		urequire.NoError(t, WriteUser(u))
		uassert.Error(t, WriteUser(u), ErrAlreadyExists.Error())
	})

	t.Run("nil user", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(std.NewCodeRealm(gnolandUsers))

		var u susers.User
		uassert.Error(t, WriteUser(u), ErrUserIsNil.Error())
	})

	t.Run("empty username", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(std.NewCodeRealm(gnolandUsers))

		u := susers.User{
			Name:    "",
			Address: alice,
		}

		uassert.Error(t, WriteUser(u), ErrEmptyUsername.Error())
	})

	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()
		std.TestSetRealm(std.NewCodeRealm(gnolandUsers))

		uassert.Error(t, WriteUser(susers.User{
			Name:    "test123",
			Address: "invalidaddress",
		}), ErrInvalidAddress.Error())

		uassert.Error(t, WriteUser(susers.User{
			Name:    "test123",
			Address: "",
		}), ErrInvalidAddress.Error())

		uassert.Error(t, WriteUser(susers.User{
			Name:    "test123",
			Address: "000000000000000000000000000000000000000000000000",
		}), ErrInvalidAddress.Error())
	})
}
