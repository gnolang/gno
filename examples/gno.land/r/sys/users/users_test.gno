package users

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	susers "gno.land/p/sys/users"
)

func TestResolveName(t *testing.T) {
	t.Run("invalid username", func(t *testing.T) {
		t.Parallel()
		uassert.Equal(t, ResolveName(""), nil)
	})

	t.Run("non-registered username", func(t *testing.T) {
		t.Parallel()
		uassert.Equal(t, ResolveName("notregistered123"), nil)
	})

	t.Run("valid username", func(t *testing.T) {
		t.Parallel()

		std.TestSetOrigCaller(std.NewCodeRealm(gnolandUsers).Addr())

		addr := testutils.TestAddress("newUsername")
		u := &susers.User{
			Name:    "newUsername", // should be aliceUsername, but realm state is not reset
			Address: addr,
		}

		urequire.NoError(t, WriteUser(u))
		uassert.Equal(t, ResolveName("newUsername").Name, "newUsername")
		uassert.Equal(t, ResolveName("newUsername").Address, addr)
		// XXX: ptr for uassert.Equal not supported yet
	})
}

func TestResolveAddress(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		t.Parallel()
		uassert.Equal(t, ResolveAddress(""), nil)
	})

	t.Run("valid address, not registered", func(t *testing.T) {
		t.Parallel()
		uassert.Equal(t, ResolveAddress(testutils.TestAddress("notregistered")), nil)
	})

	t.Run("valid registered address", func(t *testing.T) {
		t.Parallel()

		username := "alice123"
		uassert.Equal(t, ResolveAddress(alice).Name, username)
		uassert.Equal(t, ResolveAddress(alice).Address, alice)
	})
}
