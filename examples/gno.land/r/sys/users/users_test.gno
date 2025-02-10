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
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res, isLatest := ResolveName(alice)
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
		uassert.True(t, isLatest)
	})

	t.Run("name+Alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		data, _ := ResolveName(alice)
		urequire.NoError(t, data.UpdateName("alice1"))

		res, isLatest := ResolveName("alice1")
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, "alice1", res.Name())
		uassert.True(t, isLatest)
	})

	t.Run("multiple_aliases", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// RegisterUser and check each Alias
		var names []string
		names = append(names, alice)
		for i := 0; i < 5; i++ {
			alias := "alice" + strconv.Itoa(i)
			names = append(names, alias)

			data, _ := ResolveName(alice)
			urequire.NoError(t, data.UpdateName(alias))
		}

		for _, alias := range names {
			res, _ := ResolveName(alias)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.Addr())
			uassert.Equal(t, "alice4", res.Name())
		}
	})
}

func TestResolveAddress(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("single_name", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res := ResolveAddress(aliceAddr)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
	})

	t.Run("name+Alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		data, _ := ResolveName(alice)
		urequire.NoError(t, data.UpdateName("alice1"))

		res := ResolveAddress(aliceAddr)
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, "alice1", res.Name())
	})

	t.Run("multiple_aliases", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// RegisterUser and check each Alias
		var names []string
		names = append(names, alice)

		for i := 0; i < 5; i++ {
			alias := "alice" + strconv.Itoa(i)
			names = append(names, alias)
			data, _ := ResolveName(alice)
			urequire.NoError(t, data.UpdateName(alias))
		}

		res := ResolveAddress(aliceAddr)
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, "alice4", res.Name())
	})
}
