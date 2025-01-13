package users

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

var (
	alice     = "alice"
	aliceAddr = testutils.TestAddress(alice)
	bob       = "bob"
	bobAddr   = testutils.TestAddress(bob)

	whitelistedCallerAddr = std.DerivePkgAddr(gnolandUsers)
	adminRealm            = std.NewUserRealm(adminAddr)
)

func TestRegister(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("valid_registration", func(t *testing.T) {

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res, _ := ResolveName(alice)
		uassert.Equal(t, aliceAddr, res.Addr)

		res, _ = ResolveAddress(aliceAddr)
		uassert.Equal(t, alice, res.Username)
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		cleanStore(t)

		uassert.ErrorContains(t, RegisterUser("   ", aliceAddr), ErrEmptyUsername.Error())
		uassert.ErrorContains(t,
			RegisterUser("65letterusername65letterusername65letterusername65letterusername0",
				aliceAddr),
			ErrNameTooLong.Error())

		uassert.ErrorContains(t, RegisterUser(alice, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t, RegisterUser(alice, "invalidaddress"), ErrInvalidAddress.Error())
	})

	t.Run("addr_already_registered", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// Try registering again
		uassert.ErrorContains(t, RegisterUser("othername", aliceAddr), ErrAlreadyHasName.Error())
	})

	t.Run("name_taken", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// Try registering alice's name with bob's address
		uassert.ErrorContains(t, RegisterUser(alice, bobAddr), ErrNameTaken.Error())
	})

	t.Run("user_deleted", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, Delete(aliceAddr))

		// Try re-registering after deletion
		uassert.ErrorContains(t, RegisterUser("newname", aliceAddr), ErrDeletedUser.Error())
	})

	t.Run("address_lookalike", func(t *testing.T) {
		cleanStore(t)

		// Address as Username
		uassert.ErrorContains(t, RegisterUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", aliceAddr), ErrNameLikeAddress.Error())
		// Beginning of address as Username
		uassert.ErrorContains(t, RegisterUser("g1jg8mtutu9khhfwc4nxmu", aliceAddr), ErrNameLikeAddress.Error())
		uassert.NoError(t, RegisterUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5longerthananaddress", aliceAddr))
	})
}

func TestRegisterAlias(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)
	t.Run("valid_direct_alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, RegisterAlias("alice1", aliceAddr))
	})

	t.Run("valid_double_alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, RegisterAlias("alice1", aliceAddr))
		uassert.NoError(t, RegisterAlias("alice2", aliceAddr))

	})

	t.Run("name_taken", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, RegisterAlias("alice1", aliceAddr))
	})

	t.Run("alias_before_name", func(t *testing.T) {
		cleanStore(t)

		uassert.ErrorContains(t, RegisterAlias(alice, aliceAddr), ErrAliasBeforeName.Error())
	})

	t.Run("alias_after_delete", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, Delete(aliceAddr))

		uassert.ErrorContains(t, RegisterAlias("newalice", aliceAddr), ErrDeletedUser.Error())
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		uassert.ErrorContains(t, RegisterAlias("   ", aliceAddr), ErrEmptyUsername.Error())
		uassert.ErrorContains(t,
			RegisterAlias("65letterusername65letterusername65letterusername65letterusername0",
				aliceAddr),
			ErrNameTooLong.Error())

		uassert.ErrorContains(t, RegisterAlias(alice, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t, RegisterAlias(alice, "invalidaddress"), ErrInvalidAddress.Error())
	})

	t.Run("address_lookalike", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// Address as Username
		uassert.ErrorContains(t, RegisterAlias("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", aliceAddr), ErrNameLikeAddress.Error())
		// Beginning of address as Username
		uassert.ErrorContains(t, RegisterAlias("g1jg8mtutu9khhfwc4nxmu", aliceAddr), ErrNameLikeAddress.Error())
		uassert.NoError(t, RegisterAlias("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5longerthananaddress", aliceAddr))
	})
}

func TestResolveName(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("single_name", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res, deleted := ResolveName(alice)

		uassert.Equal(t, aliceAddr, res.Addr)
		uassert.Equal(t, alice, res.Username)
		uassert.Equal(t, nil, res.Alias)
		uassert.False(t, deleted)
	})

	t.Run("name+Alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, RegisterAlias("alice1", aliceAddr))

		res, deleted := ResolveName(alice)
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.Addr)
		uassert.Equal(t, "alice1", res.Username)
		uassert.Equal(t, nil, res.Alias)
		uassert.False(t, deleted)
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

			urequire.NoError(t, RegisterAlias(alias, aliceAddr))

			// Right after registering, it should resolve to itself
			res, deleted := ResolveName(alias)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.Addr)
			uassert.Equal(t, alias, res.Username)
			uassert.Equal(t, nil, res.Alias)
			uassert.False(t, deleted)
		}

		// Now verify that all aliases resolve to the latest one (alice4)
		finalAlias := names[len(names)-1]

		for _, alias := range names {
			res, deleted := ResolveName(alias)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.Addr)
			uassert.Equal(t, finalAlias, res.Username)
			uassert.Equal(t, nil, res.Alias)
			uassert.False(t, deleted)
		}
	})
}

func TestResolveAddress(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("single_name", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res, deleted := ResolveAddress(aliceAddr)

		uassert.Equal(t, aliceAddr, res.Addr)
		uassert.Equal(t, alice, res.Username)
		uassert.Equal(t, nil, res.Alias)
		uassert.False(t, deleted)
	})

	t.Run("name+Alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, RegisterAlias("alice1", aliceAddr))

		res, deleted := ResolveAddress(aliceAddr)
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.Addr)
		uassert.Equal(t, "alice1", res.Username)
		uassert.Equal(t, nil, res.Alias)
		uassert.False(t, deleted)
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

			urequire.NoError(t, RegisterAlias(alias, aliceAddr))

			// Right after registering, it should resolve to itself
			res, deleted := ResolveAddress(aliceAddr)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.Addr)
			uassert.Equal(t, alias, res.Username)
			uassert.Equal(t, nil, res.Alias)
			uassert.False(t, deleted)
		}
	})
}

func TestDelete(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("non_existent_user", func(t *testing.T) {
		cleanStore(t)

		uassert.ErrorContains(t, Delete(testutils.TestAddress("unregistered")), ErrUserNotExist.Error())
	})

	t.Run("double_delete", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, Delete(aliceAddr))
		uassert.ErrorContains(t, Delete(aliceAddr), ErrAlreadyDeleted.Error())
	})

	t.Run("valid_delete", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, Delete(aliceAddr))

		resolved, deleted := ResolveName(alice)
		uassert.NotEqual(t, nil, resolved)
		uassert.Equal(t, alice, resolved.Username)
		uassert.Equal(t, aliceAddr, resolved.Addr)
		uassert.True(t, deleted)
	})
}

// cleanStore should not be needed, as vm store should be reset after each test.
// Reference: https://github.com/gnolang/gno/issues/1982
func cleanStore(t *testing.T) {
	t.Helper()

	nameStore = avl.NewTree()
	addressStore = avl.NewTree()
}
