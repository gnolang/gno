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
		t.Parallel()

		urequire.NoError(t, Register(alice, aliceAddr))

		res := ResolveName(alice)
		uassert.Equal(t, aliceAddr, res.addr)

		res = ResolveAddress(aliceAddr)
		uassert.Equal(t, alice, res.username)
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		uassert.ErrorContains(t, Register("   ", aliceAddr), ErrEmptyUsername.Error())
		uassert.ErrorContains(t,
			Register("65letterusername65letterusername65letterusername65letterusername0",
				aliceAddr),
			ErrNameTooLong.Error())

		uassert.ErrorContains(t, Register(alice, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t, Register(alice, "invalidaddress"), ErrInvalidAddress.Error())
	})

	t.Run("addr_already_registered", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))

		// Try registering again
		uassert.ErrorContains(t, Register("othername", aliceAddr), ErrAlreadyHasName.Error())
	})

	t.Run("name_taken", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))

		// Try registering alice's name with bob's address
		uassert.ErrorContains(t, Register(alice, bobAddr), ErrNameTaken.Error())
	})

	t.Run("user_deleted", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		urequire.NoError(t, Delete(alice))

		// Try re-registering after deletion
		uassert.ErrorContains(t, Register("newname", aliceAddr), ErrDeletedUser.Error())
	})
}

func TestRegisterAlias(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)
	t.Run("valid_direct_alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		uassert.NoError(t, RegisterAlias("alice1", aliceAddr))
	})

	t.Run("valid_double_alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		uassert.NoError(t, RegisterAlias("alice1", aliceAddr))
		uassert.NoError(t, RegisterAlias("alice2", aliceAddr))

	})

	t.Run("name_taken", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		uassert.NoError(t, RegisterAlias("alice1", aliceAddr))
	})

	t.Run("alias_before_name", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		uassert.ErrorContains(t, RegisterAlias(alice, aliceAddr), ErrAliasBeforeName.Error())
	})

	t.Run("alias_after_delete", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		urequire.NoError(t, Delete(alice))

		uassert.ErrorContains(t, RegisterAlias("newalice", aliceAddr), ErrDeletedUser.Error())
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))

		uassert.ErrorContains(t, RegisterAlias("   ", aliceAddr), ErrEmptyUsername.Error())
		uassert.ErrorContains(t,
			RegisterAlias("65letterusername65letterusername65letterusername65letterusername0",
				aliceAddr),
			ErrNameTooLong.Error())

		uassert.ErrorContains(t, RegisterAlias(alice, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t, RegisterAlias(alice, "invalidaddress"), ErrInvalidAddress.Error())
	})
}

func TestResolveName(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("single_name", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))

		res := ResolveName(alice)

		uassert.Equal(t, aliceAddr, res.addr)
		uassert.Equal(t, alice, res.username)
		uassert.Equal(t, nil, res.alias)
		uassert.False(t, res.deleted)
	})

	t.Run("name+alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		urequire.NoError(t, RegisterAlias("alice1", aliceAddr))

		res := ResolveName(alice)
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.addr)
		uassert.Equal(t, "alice1", res.username)
		uassert.Equal(t, nil, res.alias)
		uassert.False(t, res.deleted)
	})

	t.Run("multiple_aliases", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))

		// Register and check each alias
		var names []string
		names = append(names, alice)
		for i := 0; i < 5; i++ {
			alias := "alice" + strconv.Itoa(i)
			names = append(names, alias)

			urequire.NoError(t, RegisterAlias(alias, aliceAddr))

			// Right after registering, it should resolve to itself
			res := ResolveName(alias)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.addr)
			uassert.Equal(t, alias, res.username)
			uassert.Equal(t, nil, res.alias)
			uassert.False(t, res.deleted)
		}

		// Now verify that all aliases resolve to the latest one (alice4)
		finalAlias := names[len(names)-1]

		for _, alias := range names {
			res := ResolveName(alias)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.addr)
			uassert.Equal(t, finalAlias, res.username)
			uassert.Equal(t, nil, res.alias)
			uassert.False(t, res.deleted)
		}
	})
}

func TestResolveAddress(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("single_name", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))

		res := ResolveAddress(aliceAddr)

		uassert.Equal(t, aliceAddr, res.addr)
		uassert.Equal(t, alice, res.username)
		uassert.Equal(t, nil, res.alias)
		uassert.False(t, res.deleted)
	})

	t.Run("name+alias", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		urequire.NoError(t, RegisterAlias("alice1", aliceAddr))

		res := ResolveAddress(aliceAddr)
		urequire.NotEqual(t, nil, res)

		uassert.Equal(t, aliceAddr, res.addr)
		uassert.Equal(t, "alice1", res.username)
		uassert.Equal(t, nil, res.alias)
		uassert.False(t, res.deleted)
	})

	t.Run("multiple_aliases", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))

		// Register and check each alias
		var names []string
		names = append(names, alice)

		for i := 0; i < 5; i++ {
			alias := "alice" + strconv.Itoa(i)
			names = append(names, alias)

			urequire.NoError(t, RegisterAlias(alias, aliceAddr))

			// Right after registering, it should resolve to itself
			res := ResolveAddress(aliceAddr)
			urequire.NotEqual(t, nil, res)

			uassert.Equal(t, aliceAddr, res.addr)
			uassert.Equal(t, alias, res.username)
			uassert.Equal(t, nil, res.alias)
			uassert.False(t, res.deleted)
		}
	})
}

func TestDelete(t *testing.T) {
	std.TestSetOrigCaller(whitelistedCallerAddr)

	t.Run("non_existent_user", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		uassert.ErrorContains(t, Delete("someone"), ErrUserNotExist.Error())
	})

	t.Run("double_delete", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		urequire.NoError(t, Delete(alice))

		uassert.ErrorContains(t, Delete(alice), ErrAlreadyDeleted.Error())
	})

	t.Run("valid_delete", func(t *testing.T) {
		t.Parallel()
		cleanStore(t)

		urequire.NoError(t, Register(alice, aliceAddr))
		uassert.NoError(t, Delete(alice))

		resolved := ResolveName(alice)
		uassert.NotEqual(t, nil, resolved)
		uassert.Equal(t, alice, resolved.username)
		uassert.Equal(t, aliceAddr, resolved.addr)
		uassert.True(t, resolved.deleted)
	})
}

// cleanStore should not be needed, as vm store should be reset after each test.
// Reference: https://github.com/gnolang/gno/issues/1982
func cleanStore(t *testing.T) {
	t.Helper()

	nameStore = avl.NewTree()
	addressStore = avl.NewTree()
}
