package names

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"

	"gno.land/r/sys/users"
)

var alice = testutils.TestAddress("alice")

func TestDefaultVerifier(t *testing.T) {
	// Check disabled, any case is true
	uassert.True(t, verifier(false, alice, alice.String()))
	uassert.True(t, verifier(false, "", alice.String()))
	uassert.True(t, verifier(false, alice, "somerandomusername"))

	// Check enabled
	// username + addr mismatch
	uassert.False(t, verifier(true, alice, "notregistered"))
	// PA namespace check
	uassert.True(t, verifier(true, alice, alice.String()))

	// Empty name/address
	uassert.False(t, verifier(true, std.Address(""), ""))

	// Register proper username
	testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoland/users/v1")) // authorized write
	urequire.NoError(t, cross(users.RegisterUser)("alice", alice))

	// Proper namespace
	uassert.True(t, verifier(true, alice, "alice"))
}

func TestEnable(t *testing.T) {
	testing.SetRealm(std.NewUserRealm("g1manfred47kzduec920z88wfr64ylksmdcedlf5"))

	uassert.NotAborts(t, func() {
		cross(Enable)()
	})

	// Confirm enable drops ownerships
	uassert.Equal(t, Ownable.Owner().String(), "")
	uassert.AbortsWithMessage(t, ownable.ErrUnauthorized.Error(), func() {
		cross(Enable)()
	})
}

func TestVerify(t *testing.T) {
	t.Run("verify from unauthorized", func(t *testing.T) {
		std.TestSetOrigCaller(acc2) // acc2 is not the owner

		uassert.AbortsWithMessage(t, ownable.ErrUnauthorized.Error(), func() {
			v.Verify(name, "test-verification")
		})
	})

	t.Run("verify from owner", func(t *testing.T) {
		std.TestSetOrigCaller(owner)

		uassert.NotAborts(t, func() {
			v.Verify(name, "test-verification")
		})

		verifiedName, err := v.GetVerifiedName(owner)
		// ... existing code ...
	})
}
