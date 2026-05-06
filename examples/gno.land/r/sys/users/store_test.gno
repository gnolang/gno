package users

import (
	"chain"
	"testing"

	"gno.land/p/nt/bptree/v0"
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

		// Lowercase-only enforcement (closes case-confusable squatting via
		// Alice/alice/ALICE registering as distinct names — see TO_REVIEW
		// adversarial review). reName mirrors gno's mempackage Re_name shape.
		uassert.ErrorContains(t, RegisterUser(cross, "Alice", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "ALICE", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "aLice", aliceAddr), ErrInvalidUsername.Error())

		// Must start with a lowercase letter. Names starting with a digit,
		// underscore, or hyphen are rejected.
		uassert.ErrorContains(t, RegisterUser(cross, "1alice", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "9", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "_alice", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "-alice", aliceAddr), ErrInvalidUsername.Error())

		// Cannot end with a separator either (the new mempackage-aligned
		// regex requires the final char to be in [a-z0-9]).
		uassert.ErrorContains(t, RegisterUser(cross, "alice_", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "alice-", aliceAddr), ErrInvalidUsername.Error())

		// Cannot have consecutive separators in the middle either.
		// Each separator MUST be followed by at least one alphanumeric.
		uassert.ErrorContains(t, RegisterUser(cross, "alice--bob", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "alice__bob", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "alice-_bob", aliceAddr), ErrInvalidUsername.Error())
		uassert.ErrorContains(t, RegisterUser(cross, "alice_-bob", aliceAddr), ErrInvalidUsername.Error())

		// Length cap of exactly 64 — boundary cases.
		exactly65 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // 65 chars
		uassert.ErrorContains(t, RegisterUser(cross, exactly65, aliceAddr), ErrInvalidUsername.Error())
	})

	t.Run("valid_edge_cases", func(t *testing.T) {
		// Names valid under the mempackage-aligned reName: starts with a
		// letter, may contain hyphens or underscores in the middle, ends
		// in [a-z0-9]. The hyphen support is what enables the namereg/v1
		// `nym-...` prefix shape.
		cleanStore(t)
		urequire.NoError(t, RegisterUser(cross,
			"nym-alice123",
			testutils.TestAddress("u_nym_alice")))

		cleanStore(t)
		urequire.NoError(t, RegisterUser(cross,
			"a_b_c",
			testutils.TestAddress("u_underscore")))

		cleanStore(t)
		urequire.NoError(t, RegisterUser(cross,
			"a",
			testutils.TestAddress("u_singlechar")))

		cleanStore(t)
		// 64 chars exactly (length cap is inclusive).
		exactly64 := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		urequire.NoError(t, RegisterUser(cross,
			exactly64,
			testutils.TestAddress("u_64")))
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

func TestCanonicalize(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		// Single-rule cases.
		{"l", "i"}, {"i", "i"}, {"1", "i"},
		{"0", "o"}, {"o", "o"},
		{"-", ""}, {".", ""}, {"_", ""},

		// Identity / pass-through.
		{"abc", "abc"}, {"xyz", "xyz"}, {"123", "i23"},

		// Open Nym Tier examples from the design doc.
		{"nym-vital1k123", "nymvitaiiki23"},
		{"nym-vitalik123", "nymvitaiiki23"},
		{"nym-vital1k999", "nymvitaiik999"},

		// Same stem, different digit suffix → distinct canonicals.
		{"nym-foolbar000", "nymfooibarooo"},
		{"nym-foolbar001", "nymfooibarooi"},

		// Three-way separator equivalence.
		{"xyz-com", "xyzcom"},
		{"xyz_com", "xyzcom"},
		{"xyz.com", "xyzcom"},

		// Idempotency: applying twice equals applying once.
		{"nymvitaiiki23", "nymvitaiiki23"},

		// Empty input.
		{"", ""},
	}
	for _, tc := range cases {
		got := Canonicalize(tc.in)
		uassert.Equal(t, tc.want, got)
	}
}

func TestCanonicalize_NonASCIIPassthrough(t *testing.T) {
	// ASCII-only contract: the function passes non-ASCII bytes through
	// unchanged. Locks the documented contract — registration paths
	// reject non-ASCII upstream via validateName, so this code path is
	// only reachable through direct callers from other realms.
	uassert.Equal(t, "café", Canonicalize("café"))
	uassert.Equal(t, "naïve", Canonicalize("naïve"))
}

func TestRegisterUser_CanonicalCollision(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

	t.Run("non_bypass_blocks", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))
		// Different exact name, same canonical form → ErrCanonicalCollision.
		uassert.ErrorContains(t,
			RegisterUser(cross, "vital1k", bobAddr),
			ErrCanonicalCollision.Error(),
		)
	})

	t.Run("non_bypass_blocks_with_separator_strip", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, "xyz-com", aliceAddr))
		uassert.ErrorContains(t,
			RegisterUser(cross, "xyz_com", bobAddr),
			ErrCanonicalCollision.Error(),
		)
	})

	t.Run("non_bypass_allows_different_digit_suffix", func(t *testing.T) {
		cleanStore(t)

		// Same alpha stem, different digits-after-canonicalization → no collision.
		urequire.NoError(t, RegisterUser(cross, "nym-foolbar000", aliceAddr))
		urequire.NoError(t, RegisterUser(cross, "nym-foolbar999",
			testutils.TestAddress("foolbar999")))
	})

	t.Run("exact_name_taken_takes_precedence", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))
		// Same exact name → ErrNameTaken (NOT ErrCanonicalCollision).
		// The nameStore.Has check precedes the canonical check.
		uassert.ErrorContains(t,
			RegisterUser(cross, "vitalik", bobAddr),
			ErrNameTaken.Error(),
		)
	})
}

func TestRegisterUserIgnoreCanonical(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

	t.Run("bypass_succeeds_on_canonical_collision", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))
		// Bypass succeeds even though canonical collides.
		urequire.NoError(t, RegisterUserIgnoreCanonical(cross, "vital1k", bobAddr))

		// Both names resolve.
		res, _ := ResolveName("vitalik")
		uassert.Equal(t, aliceAddr, res.Addr())
		res, _ = ResolveName("vital1k")
		uassert.Equal(t, bobAddr, res.Addr())
	})

	t.Run("bypass_later_wins_overwrite", func(t *testing.T) {
		cleanStore(t)

		// First registration writes canonical entry.
		urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))
		existing, taken := IsCanonicalTaken("vital1k")
		urequire.True(t, taken)
		uassert.Equal(t, "vitalik", existing)

		// Bypass write overwrites the canonical entry (decision #14).
		urequire.NoError(t, RegisterUserIgnoreCanonical(cross, "vital1k", bobAddr))
		existing, taken = IsCanonicalTaken("vitalik")
		urequire.True(t, taken)
		uassert.Equal(t, "vital1k", existing)
	})

	t.Run("bypass_still_blocks_exact_taken", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))
		// Same exact name still blocked, even via the bypass path.
		uassert.ErrorContains(t,
			RegisterUserIgnoreCanonical(cross, "vitalik", bobAddr),
			ErrNameTaken.Error(),
		)
	})
}

func TestUpdateName_CanonicalCollision(t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))

	t.Run("self_collision_blocks_per_decision_15", func(t *testing.T) {
		cleanStore(t)

		// Alice registers vital1k. Canonical store maps "vitaiik" → "vital1k".
		urequire.NoError(t, RegisterUser(cross, "vital1k", aliceAddr))
		data := ResolveAddress(aliceAddr)

		// Alice attempts to rename to a confusable variant of HER OWN name.
		// Decision #15: blocked. No self-collision exception in the
		// non-bypass path.
		testing.SetOriginCaller(whitelistedCallerAddr)
		uassert.ErrorContains(t,
			data.UpdateName("vitalik"),
			ErrCanonicalCollision.Error(),
		)
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
	})

	t.Run("different_user_collision_blocks", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))
		urequire.NoError(t, RegisterUser(cross, "bob", bobAddr))

		data := ResolveAddress(bobAddr)
		testing.SetOriginCaller(whitelistedCallerAddr)
		uassert.ErrorContains(t,
			data.UpdateName("vital1k"),
			ErrCanonicalCollision.Error(),
		)
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))
	})

	t.Run("bypass_allows_self_collision_rename", func(t *testing.T) {
		cleanStore(t)

		urequire.NoError(t, RegisterUser(cross, "vital1k", aliceAddr))
		data := ResolveAddress(aliceAddr)

		// Bypass path allows the rename (DAO grant scenario).
		testing.SetOriginCaller(whitelistedCallerAddr)
		urequire.NoError(t, data.UpdateNameIgnoreCanonical("vitalik"))
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))

		// Both names point to alice; latest is "vitalik".
		uassert.Equal(t, "vitalik", ResolveAddress(aliceAddr).Name())
		// Old name still resolves to alice (decision #4: keep old entry).
		res, isLatest := ResolveName("vital1k")
		uassert.Equal(t, aliceAddr, res.Addr())
		uassert.False(t, isLatest)
	})
}

func TestCanonicalEntry_Persists_After_Delete(t *testing.T) {
	// Decision #5: Delete does NOT remove the canonical entry. Mirrors
	// the existing tombstone behavior (anti-revival policy).
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))
	cleanStore(t)

	urequire.NoError(t, RegisterUser(cross, "vitalik", aliceAddr))
	data := ResolveAddress(aliceAddr)
	urequire.NoError(t, data.Delete())

	// Canonical entry retained.
	existing, taken := IsCanonicalTaken("vitalik")
	urequire.True(t, taken)
	uassert.Equal(t, "vitalik", existing)

	// A different user CANNOT register a confusable variant — the deleted
	// user's canonical claim still stands.
	uassert.ErrorContains(t,
		RegisterUser(cross, "vital1k", bobAddr),
		ErrCanonicalCollision.Error(),
	)
}

func TestCanonicalEntry_Persists_After_UpdateName(t *testing.T) {
	// Decision #4: UpdateName keeps the OLD canonical entry. Mirrors the
	// existing nameStore alias retention (anti-rename-squat policy).
	testing.SetRealm(testing.NewCodeRealm(initControllerPath))
	cleanStore(t)

	urequire.NoError(t, RegisterUser(cross, "alice", aliceAddr))
	data := ResolveAddress(aliceAddr)
	testing.SetOriginCaller(whitelistedCallerAddr)
	urequire.NoError(t, data.UpdateName("alice2"))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/sys/users"))

	// Both old and new canonical entries exist.
	existing, taken := IsCanonicalTaken("alice")
	urequire.True(t, taken)
	uassert.Equal(t, "alice", existing)

	existing, taken = IsCanonicalTaken("alice2")
	urequire.True(t, taken)
	uassert.Equal(t, "alice2", existing)
}

// cleanStore should not be needed, as vm store should be reset after each test.
// Reference: https://github.com/gnolang/gno/issues/1982
func cleanStore(t *testing.T) {
	t.Helper()

	nameStore = bptree.NewBPTree32()
	addressStore = bptree.NewBPTree32()
	canonicalStore = bptree.NewBPTree32()
}
