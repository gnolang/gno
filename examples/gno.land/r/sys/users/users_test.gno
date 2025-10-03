package users

import (
	"strconv"
	"testing"

	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
)

func TestResolveName(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(gUsersV1Path))

	t.Run("single_name", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		res, isLatest := ResolveName(alice)
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
		uassert.True(t, isLatest)
	})

	t.Run("name+Alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
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

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

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
	testing.SetRealm(testing.NewCodeRealm(gUsersV1Path))

	t.Run("single_name", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		res := ResolveAddress(aliceAddr)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
	})

	t.Run("name+Alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data, _ := ResolveName(alice)
		urequire.NoError(t, data.UpdateName("alice1"))

		res := ResolveAddress(aliceAddr)
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, "alice1", res.Name())
	})

	t.Run("multiple_aliases", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

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

func TestROStores(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(gUsersV1Path))
	cleanStore(t)

	urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
	roNS := GetReadOnlyNameStore()
	roAS := GetReadonlyAddrStore()

	t.Run("get user data", func(t *testing.T) {
		// Name store
		aliceDataRaw, ok := roNS.Get(alice)
		uassert.True(t, ok)

		roData, ok := aliceDataRaw.(*UserData)
		uassert.True(t, ok, "Could not cast data from RO tree to UserData")

		// Try to modify data
		roData.Delete()
		raw, ok := nameStore.Get(alice)
		uassert.False(t, raw.(*UserData).deleted)

		// Addr store
		aliceDataRaw, ok = roAS.Get(aliceAddr.String())
		uassert.True(t, ok)

		roData, ok = aliceDataRaw.(*UserData)
		uassert.True(t, ok, "Could not cast data from RO tree to UserData")

		// Try to modify data
		roData.Delete()
		raw, ok = nameStore.Get(alice)
		uassert.False(t, raw.(*UserData).deleted)
	})

	t.Run("get deleted data", func(t *testing.T) {
		raw, _ := nameStore.Get(alice)
		aliceData := raw.(*UserData)

		urequire.NoError(t, aliceData.Delete())
		urequire.True(t, aliceData.IsDeleted())

		// Should be nil because of makeSafeFn
		rawRoData, ok := roNS.Get(alice)
		// uassert.False(t, ok)
		// XXX: not sure what to do here, as the tree technically has the data so returns ok
		// However the data is intercepted and something else (nil in this case) is returned.
		// should we handle this somehow?

		uassert.Equal(t, rawRoData, nil)
		_, ok = rawRoData.(*UserData) // shouldn't be castable
		uassert.False(t, ok)
	})
}

func TestResolveAny(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(gUsersV1Path))

	t.Run("name", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		res, _ := ResolveAny(alice)

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
	})

	t.Run("address", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		res, _ := ResolveAny(aliceAddr.String())

		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.Equal(t, alice, res.Name())
	})

	t.Run("not_registered", func(t *testing.T) {
		cleanStore(t)

		res, _ := ResolveAny(aliceAddr.String())

		uassert.Equal(t, nil, res)
	})
}

// TODO Uncomment after gnoweb /u/ page.
//func TestUserRenderLink(t *testing.T) {
//	testing.SetOriginCaller(whitelistedCallerAddr)
//	cleanStore(t)
//
//	urequire.NoError(t, RegisterUser(alice, aliceAddr))
//
//	data, _ := ResolveName(alice)
//	uassert.Equal(t, data.RenderLink(""), ufmt.Sprintf("[@%s](/u/%s)", alice, alice))
//	text := "my link text!"
//	uassert.Equal(t, data.RenderLink(text), ufmt.Sprintf("[%s](/u/%s)", text, alice))
//}
