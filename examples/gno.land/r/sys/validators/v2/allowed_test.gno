package validators

import (
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/urequire"
	"gno.land/p/sys/validators"
)

func TestAllowedRealms_AddRemove(t *testing.T) {
	allowedRealms = avl.NewTree()

	uassert.False(t, IsAllowedRealm("gno.land/r/demo/ics"))

	allowedRealms.Set("gno.land/r/demo/ics", true)
	uassert.True(t, IsAllowedRealm("gno.land/r/demo/ics"))

	realms := GetAllowedRealms()
	uassert.Equal(t, 1, len(realms))
	uassert.Equal(t, "gno.land/r/demo/ics", realms[0])

	allowedRealms.Remove("gno.land/r/demo/ics")
	uassert.False(t, IsAllowedRealm("gno.land/r/demo/ics"))
	uassert.Equal(t, 0, len(GetAllowedRealms()))
}

func TestAddValidator_NotAllowed(cur realm, t *testing.T) {
	allowedRealms = avl.NewTree()
	changes = avl.NewTree()

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/unknown"))

	val := validators.Validator{
		Address:     testutils.TestAddress("val1"),
		PubKey:      "public-key",
		VotingPower: 10,
	}

	urequire.AbortsWithMessage(t,
		"realm gno.land/r/demo/unknown is not allowed to modify the validator set",
		func() { AddValidator(cross, val) },
	)
}

func TestAddValidator_UserNotAllowed(cur realm, t *testing.T) {
	allowedRealms = avl.NewTree()
	changes = avl.NewTree()

	testing.SetRealm(testing.NewUserRealm(testutils.TestAddress("user1")))

	val := validators.Validator{
		Address:     testutils.TestAddress("val1"),
		PubKey:      "public-key",
		VotingPower: 10,
	}

	urequire.AbortsWithMessage(t,
		"caller must be a realm, not a user",
		func() { AddValidator(cross, val) },
	)
}

func TestAddValidator_Allowed(cur realm, t *testing.T) {
	allowedRealms = avl.NewTree()
	changes = avl.NewTree()

	const allowedPath = "gno.land/r/demo/ics"
	allowedRealms.Set(allowedPath, true)

	testing.SetRealm(testing.NewCodeRealm(allowedPath))

	val := validators.Validator{
		Address:     testutils.TestAddress("val-allowed"),
		PubKey:      "public-key",
		VotingPower: 10,
	}

	AddValidator(cross, val)
	uassert.True(t, vp.IsValidator(val.Address))
}

func TestRemoveValidator_Allowed(cur realm, t *testing.T) {
	allowedRealms = avl.NewTree()
	changes = avl.NewTree()

	const allowedPath = "gno.land/r/demo/ics"
	allowedRealms.Set(allowedPath, true)

	testing.SetRealm(testing.NewCodeRealm(allowedPath))

	val := validators.Validator{
		Address:     testutils.TestAddress("val-remove"),
		PubKey:      "public-key",
		VotingPower: 10,
	}

	AddValidator(cross, val)
	uassert.True(t, vp.IsValidator(val.Address))

	RemoveValidator(cross, val.Address)
	uassert.False(t, vp.IsValidator(val.Address))
}

func TestRemoveValidator_NotAllowed(cur realm, t *testing.T) {
	allowedRealms = avl.NewTree()
	changes = avl.NewTree()

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/unknown"))

	urequire.AbortsWithMessage(t,
		"realm gno.land/r/demo/unknown is not allowed to modify the validator set",
		func() { RemoveValidator(cross, testutils.TestAddress("val1")) },
	)
}
