package users

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

func TestResolveName(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("single_name", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res, deleted, isAlias := ResolveName(alice)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
		uassert.False(t, deleted)
		uassert.False(t, isAlias)
	})

	t.Run("name+Alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, ChangeName("alice1", aliceAddr))

		res, deleted, isAlias := ResolveName("alice1")
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, "alice1", res.Name())
		uassert.False(t, deleted)
		uassert.False(t, isAlias)
	})

	t.Run("multiple_aliases", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// RegisterUser and check each Alias
		var names []string
		names = append(names, alice)
		for i := 0; i < 5; i++ {
			alias := "alice" + strconv.Itoa(i)
			names = append(names, alias)

			urequire.NoError(t, ChangeName(alias, aliceAddr))
		}

		for _, alias := range names {
			res, deleted, isAlias := ResolveName(alias)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.Addr())
			uassert.Equal(t, "alice4", res.Name())
			uassert.False(t, deleted)
		}
	})
}

func TestResolveAddress(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("single_name", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res, deleted := ResolveAddress(aliceAddr)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
		uassert.False(t, deleted)
	})

	t.Run("name+Alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, ChangeName("alice1", aliceAddr))

		res, deleted := ResolveAddress(aliceAddr)
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, "alice1", res.Name())
		uassert.False(t, deleted)
	})

	t.Run("multiple_aliases", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// RegisterUser and check each Alias
		var names []string
		names = append(names, alice)

		for i := 0; i < 5; i++ {
			alias := "alice" + strconv.Itoa(i)
			names = append(names, alias)
			urequire.NoError(t, ChangeName(alias, aliceAddr))
		}

		res, deleted := ResolveAddress(aliceAddr)
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, "alice4", res.Name())
		uassert.False(t, deleted)
	})
}
