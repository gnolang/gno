package users

import (
	"chain"
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	alice     = "alice"
	aliceAddr = testutils.TestAddress(alice)
	bob       = "bob"
	bobAddr   = testutils.TestAddress(bob)

	whitelistedCallerAddr = chain.PackageAddress(initControllerPath)
)

func TestRegister(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

	t.Run("valid_registration", func(t *testing.T) {
		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		res, isLatest := ResolveName(alice)
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.True(t, isLatest)

		res = ResolveAddress(aliceAddr)
		uassert.Equal(t, alice, res.Name())
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		cleanStore(t)

		uassert.ErrorContains(t, RegisterUser(cross, "", aliceAddr), ErrEmptyUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, alice, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t, RegisterUser(cross, alice, "invalidaddress"), ErrInvalidAddress.Error())

		uassert.ErrorContains(t, RegisterUser(cross, "username with a space", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t,
			RegisterUser(cross, "verylongusernameverylongusernameverylongusernameverylongusername1", aliceAddr),
			ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "namewith^&()", aliceAddr), ErrInvalidUsername.Error())
	})

	t.Run("addr_already_registered", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		// Try registering again
		uassert.ErrorContains(t, RegisterUser(cross, "othername", aliceAddr), ErrAlreadyHasName.Error())
	})

	t.Run("name_taken", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		// Try registering alice's name with bob's address
		uassert.ErrorContains(t, RegisterUser(cross, alice, bobAddr), ErrNameTaken.Error())
	})

	t.Run("user_deleted", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)
		urequire.NoError(t, data.Delete())

		// Try re-registering after deletion
		uassert.ErrorContains(t, RegisterUser(cross, "newname", aliceAddr), ErrDeletedUser.Error())
	})

	t.Run("address_lookalike", func(t *testing.T) {
		cleanStore(t)

		// Address as username
		uassert.ErrorContains(t, RegisterUser(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", aliceAddr), ErrNameLikeAddress.Error())
		// Beginning of address as username
		uassert.ErrorContains(t, RegisterUser(cross, "g1jg8mtutu9khhfwc4nxmu", aliceAddr), ErrNameLikeAddress.Error())
		uassert.NoError(t, RegisterUser(cross, "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5longerthananaddress", aliceAddr))
	})
}

func TestUpdateName(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

	t.Run("valid_direct_alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)
		{
			testing.SetOriginCaller(whitelistedCallerAddr)
			uassert.NoError(t, data.UpdateName("alice1"))
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}
	})

	t.Run("valid_double_alias", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)
		{
			testing.SetOriginCaller(whitelistedCallerAddr)
			uassert.NoError(t, data.UpdateName("alice2"))
			uassert.NoError(t, data.UpdateName("alice3"))
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}
		uassert.Equal(t, ResolveAddress(aliceAddr).username, "alice3")
	})

	t.Run("name_taken", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))

		data := ResolveAddress(aliceAddr)
		uassert.Error(t, data.UpdateName(alice), ErrNameTaken.Error())
	})

	t.Run("alias_before_name", func(t *testing.T) {
		cleanStore(t)
		data := ResolveAddress(aliceAddr) // not registered

		uassert.ErrorContains(t, data.UpdateName(alice), ErrUserNotExistOrDeleted.Error())
	})

	t.Run("alias_after_delete", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)
		{
			urequire.NoError(t, data.Delete())
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}

		data = ResolveAddress(aliceAddr)
		{
			uassert.ErrorContains(t, data.UpdateName("newalice"), ErrUserNotExistOrDeleted.Error())
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}
	})

	// Audit finding #3: a controller holding a cached *UserData pointer to a
	// user that gets deleted between resolve and update must not be able to
	// insert a new alias. The alias_after_delete test above re-resolves and
	// gets nil, so it only exercises the u==nil branch. This test holds the
	// pointer across the delete and verifies the deleted-flag branch.
	t.Run("alias_with_cached_pointer_after_delete", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		// Cache the pointer BEFORE deletion (simulating a controller that
		// resolved earlier and held the reference).
		cached := ResolveAddress(aliceAddr)
		urequire.NotEqual(t, nil, cached)

		// Delete via a fresh resolve.
		data := ResolveAddress(aliceAddr)
		{
			urequire.NoError(t, data.Delete())
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}

		// The cached pointer is non-nil but its .deleted is now true.
		urequire.NotEqual(t, nil, cached)
		uassert.True(t, cached.IsDeleted())

		// Attempting UpdateName on the cached pointer must reject — without
		// this check, "squattedname" would be inserted into nameStore
		// pointing at the deleted user, becoming permanently unresolvable
		// AND unregisterable.
		{
			testing.SetOriginCaller(whitelistedCallerAddr)
			uassert.ErrorContains(t, cached.UpdateName("squattedname"), ErrUserNotExistOrDeleted.Error())
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}

		// Confirm the squat didn't happen: nameStore should NOT have it.
		_, ok := nameStore.Get("squattedname")
		uassert.False(t, ok)
	})

	t.Run("invalid_inputs", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)
		{
			testing.SetOriginCaller(whitelistedCallerAddr)
			uassert.ErrorContains(t, data.UpdateName(""), ErrEmptyUsername.Error())
			uassert.ErrorContains(t, data.UpdateName("username with a space"), ErrInvalidUsername.Error())
			uassert.ErrorContains(t,
				data.UpdateName("verylongusernameverylongusernameverylongusernameverylongusername1"),
				ErrInvalidUsername.Error())
			uassert.ErrorContains(t, data.UpdateName("namewith^&()"), ErrInvalidUsername.Error())
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}
	})

	t.Run("address_lookalike", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)

		{
			// Address as username
			uassert.ErrorContains(t, data.UpdateName("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"), ErrNameLikeAddress.Error())
			// Beginning of address as username
			uassert.ErrorContains(t, data.UpdateName("g1jg8mtutu9khhfwc4nxmu"), ErrNameLikeAddress.Error())
			uassert.NoError(t, data.UpdateName("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5longerthananaddress"))
			testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
		}
	})
}

func TestDelete(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

	t.Run("non_existent_user", func(t *testing.T) {
		cleanStore(t)

		data := ResolveAddress(testutils.TestAddress("unregistered"))
		uassert.ErrorContains(t, data.Delete(), ErrUserNotExistOrDeleted.Error())
	})

	t.Run("double_delete", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)
		urequire.NoError(t, data.Delete())
		data = ResolveAddress(aliceAddr)
		uassert.ErrorContains(t, data.Delete(), ErrUserNotExistOrDeleted.Error())
	})

	t.Run("valid_delete", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, alice, aliceAddr))
		data := ResolveAddress(aliceAddr)
		uassert.NoError(t, data.Delete())

		resolved1, _ := ResolveName(alice)
		uassert.Equal(t, nil, resolved1)
		uassert.Equal(t, nil, ResolveAddress(aliceAddr))
	})
}

func TestRegisterNotWhitelisted(t *testing.T) {
	t.Run("register_not_whitelisted", func(t *testing.T) {
		uassert.ErrorContains(t, RegisterUser(cross, alice, aliceAddr), "does not exist in whitelist")
	})
}

// cleanStore should not be needed, as vm store should be reset after each test.
// Reference: https://github.com/gnolang/gno/issues/1982
func cleanStore(t *testing.T) {
	t.Helper()

	nameStore = avl.NewTree()
	addressStore = avl.NewTree()
}
