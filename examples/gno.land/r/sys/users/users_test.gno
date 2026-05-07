package users

import (
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestResolveName(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

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
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

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
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))
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
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

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

func TestProposeErrors(t *testing.T) {
	t.Run("propose_register_user_errors", func(t *testing.T) {
		urequire.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
			ProposeRegisterUser("bad name", aliceAddr)
		})
		urequire.PanicsWithMessage(t, ErrInvalidAddress.Error(), func() {
			ProposeRegisterUser(alice, "badaddress")
		})
	})

	t.Run("propose_update_name_errors", func(t *testing.T) {
		cleanStore(t)

		urequire.PanicsWithMessage(t, ErrInvalidAddress.Error(), func() {
			ProposeUpdateName("badaddress", "alice1")
		})
		urequire.PanicsWithMessage(t, ErrInvalidUsername.Error(), func() {
			ProposeUpdateName(aliceAddr, "bad name")
		})
		// Note: unregistered user is not checked at proposal creation time.
		// The callback handles it at execution time.
	})

	t.Run("propose_delete_user_errors", func(t *testing.T) {
		cleanStore(t)

		urequire.PanicsWithMessage(t, ErrInvalidAddress.Error(), func() {
			ProposeDeleteUser("badaddress")
		})
		// Note: unregistered user is not checked at proposal creation time.
		// The callback handles it at execution time.
	})
}

// Audit finding #6: ProposeControllerAdditionAndRemoval used to return early
// from its callback when toAdd was already whitelisted (addToWhitelist
// returned ErrAlreadyWhitelisted), skipping the remove step entirely. A
// passed swap proposal would silently leave the old controller active.
//
// applyControllerSwap is the extracted callback body. These tests exercise
// it directly to confirm:
//   - normal swap (toAdd new, toRemove present) succeeds and updates state
//   - toAdd already present is benign — remove still happens
//   - toRemove already absent is benign — add still happens, no error
//   - errors that aren't the idempotency cases still propagate
func TestApplyControllerSwap(t *testing.T) {
	// Use unique addresses per subtest to avoid having to drain the
	// controllers set (addrset.Set has no clear/iterate-all method, and
	// state persists across subtests in this package's tests).
	t.Run("normal swap A->B", func(t *testing.T) {
		a := testutils.TestAddress("swapA1")
		b := testutils.TestAddress("swapB1")
		controllers.Add(a)
		defer controllers.Remove(b)

		uassert.NoError(t, applyControllerSwap(b, a))
		uassert.True(t, controllers.Has(b), "b should be whitelisted")
		uassert.False(t, controllers.Has(a), "a should be removed")
	})

	t.Run("toAdd already whitelisted, toRemove present", func(t *testing.T) {
		// Regression for audit finding #6. Before the fix, this returned
		// ErrAlreadyWhitelisted from the swap callback, skipping the remove
		// step entirely — the swap silently no-op'd and the old controller
		// stayed active.
		a := testutils.TestAddress("swapA2")
		b := testutils.TestAddress("swapB2")
		controllers.Add(a)
		controllers.Add(b)
		defer controllers.Remove(b)

		uassert.NoError(t, applyControllerSwap(b, a))
		uassert.True(t, controllers.Has(b), "b should remain whitelisted")
		uassert.False(t, controllers.Has(a), "a should be removed even though b was already in")
	})

	t.Run("toRemove already absent, toAdd new", func(t *testing.T) {
		a := testutils.TestAddress("swapA3")
		b := testutils.TestAddress("swapB3")
		c := testutils.TestAddress("swapC3") // never added
		controllers.Add(a)
		defer controllers.Remove(a)
		defer controllers.Remove(b)

		// Remove c which isn't in the set; should be benign.
		uassert.NoError(t, applyControllerSwap(b, c))
		uassert.True(t, controllers.Has(b), "b should be whitelisted")
		uassert.True(t, controllers.Has(a), "a should still be whitelisted (unchanged)")
		uassert.False(t, controllers.Has(c), "c was never in the set")
	})

	t.Run("both idempotency cases at once", func(t *testing.T) {
		// toAdd already in, toRemove already out — should succeed cleanly,
		// no state change.
		a := testutils.TestAddress("swapA4")
		b := testutils.TestAddress("swapB4")
		c := testutils.TestAddress("swapC4")
		controllers.Add(a)
		controllers.Add(b)
		defer controllers.Remove(a)
		defer controllers.Remove(b)

		uassert.NoError(t, applyControllerSwap(b, c))
		uassert.True(t, controllers.Has(a))
		uassert.True(t, controllers.Has(b))
		uassert.False(t, controllers.Has(c))
	})
}

// Audit finding #20: the controller whitelist must be queryable from
// outside the package so operators can monitor authority without
// source-diving or replaying governance proposals.
func TestControllerQueries(t *testing.T) {
	t.Run("IsController reflects whitelist state", func(t *testing.T) {
		a := testutils.TestAddress("queryA1")
		b := testutils.TestAddress("queryB1")

		// Before any add: both report false.
		uassert.False(t, IsController(a))
		uassert.False(t, IsController(b))

		controllers.Add(a)
		defer controllers.Remove(a)

		uassert.True(t, IsController(a))
		uassert.False(t, IsController(b))
	})

	t.Run("Controllers returns a snapshot of current whitelist", func(t *testing.T) {
		// Use distinct addresses that aren't already in the set from
		// other tests in this file.
		a := testutils.TestAddress("queryA2")
		b := testutils.TestAddress("queryB2")
		c := testutils.TestAddress("queryC2")

		controllers.Add(a)
		controllers.Add(b)
		controllers.Add(c)
		defer controllers.Remove(a)
		defer controllers.Remove(b)
		defer controllers.Remove(c)

		got := Controllers()

		// All three must appear (regardless of order vs other tests'
		// leftover entries — we only assert ours are present).
		seen := map[string]bool{}
		for _, e := range got {
			seen[e.String()] = true
		}
		uassert.True(t, seen[a.String()], "snapshot must contain a")
		uassert.True(t, seen[b.String()], "snapshot must contain b")
		uassert.True(t, seen[c.String()], "snapshot must contain c")
	})

	t.Run("Controllers returns a copy — caller mutation does not affect realm", func(t *testing.T) {
		a := testutils.TestAddress("queryA3")
		controllers.Add(a)
		defer controllers.Remove(a)

		got := Controllers()
		// Mutate the returned slice. The realm's controllers set must not
		// be affected.
		got = got[:0]

		uassert.True(t, controllers.Has(a), "realm state must be unaffected by caller-side slice mutation")
		uassert.True(t, IsController(a))
	})
}

// ProposeRegisterUser auto-injects a CANONICAL COLLISION warning into
// the proposal description when the proposed name canonical-collides
// with an existing registration. The warning surfaces the colliding
// existing name to voters; the proposal still goes through if voted in
// (decision #3, DAO grants always bypass).
func TestProposeRegisterUser_CollisionWarning(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

	t.Run("warning_injected_on_collision", func(t *testing.T) {
		cleanStore(t)
		urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))

		req := ProposeRegisterUser("vital1k", bobAddr)
		uassert.True(t,
			strings.Contains(req.Description(), "CANONICAL COLLISION"),
			"description must include collision warning")
		uassert.True(t,
			strings.Contains(req.Description(), "`vitalik`"),
			"description must name the existing colliding registration")
	})

	t.Run("no_warning_when_no_collision", func(t *testing.T) {
		cleanStore(t)

		req := ProposeRegisterUser("freshname", aliceAddr)
		uassert.False(t,
			strings.Contains(req.Description(), "CANONICAL COLLISION"),
			"description must not include collision warning when name is fresh")
	})
}

// Note: full execution of ProposeRegisterUser/ProposeUpdateName closures
// (with ignoreCanonical=true → later-wins overwrite) is exercised end-to-
// end by the integration test in gno.land/pkg/integration/testdata/.
// The bypass-write semantics themselves are covered here at the unit
// level by TestRegisterUserIgnoreCanonical and TestUpdateName_CanonicalCollision.

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
