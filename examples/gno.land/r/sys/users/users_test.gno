package users

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
	pusers "gno.land/p/demo/users"
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

		username := "alice123"
		u := &pusers.User{
			Name:    username,
			Address: alice,
		}

		urequire.NoError(t, WriteUser(u))
		uassert.Equal(t, ResolveName("alice123").Name, username)
		uassert.Equal(t, ResolveName("alice123").Address, alice)
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

func TestUpdateBio(t *testing.T) {
	t.Run("caller not registered", func(t *testing.T) {
		t.Parallel()

		std.TestSetOrigCaller(testutils.TestAddress("bob"))
		uassert.Error(t, UpdateBio("newbio"), ErrNotRegistered.Error())
	})

}
