package users

import (
	"std"
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
		t.Parallel()

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		res, isLatest := ResolveName(alice)
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.True(t, isLatest)

		res = ResolveAddress(aliceAddr)
		uassert.Equal(t, alice, res.Name())
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// Try registering again
		uassert.ErrorContains(t, RegisterUser("othername", aliceAddr), ErrAlreadyHasName.Error())
	})

	t.Run("name_taken", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// Try registering alice's name with bob's address
		uassert.ErrorContains(t, RegisterUser(alice, bobAddr), ErrNameTaken.Error())
	})

	t.Run("user_deleted", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, Delete(aliceAddr))

		// Try re-registering after deletion
		uassert.ErrorContains(t, RegisterUser("newname", aliceAddr), ErrDeletedUser.Error())
	})

	t.Run("address_lookalike", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		// Address as username
		uassert.ErrorContains(t, RegisterUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", aliceAddr), ErrNameLikeAddress.Error())
		// Beginning of address as username
		uassert.ErrorContains(t, RegisterUser("g1jg8mtutu9khhfwc4nxmu", aliceAddr), ErrNameLikeAddress.Error())
		uassert.NoError(t, RegisterUser("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5longerthananaddress", aliceAddr))
	})
}

func TestChangeName(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)
	t.Run("valid_direct_alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, ChangeName("alice1", aliceAddr))
	})

	t.Run("valid_double_alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, ChangeName("alice1", aliceAddr))
		uassert.NoError(t, ChangeName("alice2", aliceAddr))
	})

	t.Run("name_taken", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, ChangeName("alice1", aliceAddr))
	})

	t.Run("alias_before_name", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		uassert.ErrorContains(t, ChangeName(alice, aliceAddr), ErrAliasBeforeName.Error())
	})

	t.Run("alias_after_delete", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		urequire.NoError(t, Delete(aliceAddr))

		uassert.ErrorContains(t, ChangeName("newalice", aliceAddr), ErrDeletedUser.Error())
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		uassert.ErrorContains(t, ChangeName("   ", aliceAddr), ErrEmptyUsername.Error())
		uassert.ErrorContains(t,
			ChangeName("65letterusername65letterusername65letterusername65letterusername0",
				aliceAddr),
			ErrNameTooLong.Error())

		uassert.ErrorContains(t, ChangeName(alice, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t, ChangeName(alice, "invalidaddress"), ErrInvalidAddress.Error())
	})

	t.Run("address_lookalike", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		// Address as username
		uassert.ErrorContains(t, ChangeName("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", aliceAddr), ErrNameLikeAddress.Error())
		// Beginning of address as username
		uassert.ErrorContains(t, ChangeName("g1jg8mtutu9khhfwc4nxmu", aliceAddr), ErrNameLikeAddress.Error())
		uassert.NoError(t, ChangeName("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5longerthananaddress", aliceAddr))
	})
}

func TestDelete(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("non_existent_user", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		uassert.ErrorContains(t, Delete(testutils.TestAddress("unregistered")), ErrUserNotExist.Error())
	})

	t.Run("double_delete", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))

		urequire.NoError(t, Delete(aliceAddr))

		uassert.ErrorContains(t, Delete(aliceAddr), ErrUserNotExist.Error())
	})

	t.Run("valid_delete", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, RegisterUser(alice, aliceAddr))
		uassert.NoError(t, Delete(aliceAddr))

		resolved, _ := ResolveName(alice)
		uassert.Equal(t, nil, resolved)
	})
}

// cleanStore should not be needed, as vm store should be reset after each test.
// Reference: https://github.com/gnolang/gno/issues/1982
func cleanStore(t *testing.T) {
	t.Helper()

	nameStore = avl.NewTree()
	addressStore = avl.NewTree()
}
