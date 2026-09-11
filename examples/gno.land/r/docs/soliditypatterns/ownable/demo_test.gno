package ownable

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestInitOwnerThenChange(cur realm, t *testing.T) {
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	testing.SetRealm(testing.NewUserRealm(alice))
	InitOwner(cross(cur))
	uassert.Equal(t, Owner().String(), alice.String(), "first caller must become the owner")

	// A second InitOwner call must be rejected.
	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.AbortsWithMessage(t, cur, "owner already defined", func() {
		InitOwner(cross(cur))
	}, "InitOwner must only succeed once")

	// A non-owner must not be able to change the owner.
	uassert.AbortsWithMessage(t, cur, "caller isn't the owner", func() {
		ChangeOwner(cross(cur), bob)
	}, "non-owner must not be able to change the owner")
	uassert.Equal(t, Owner().String(), alice.String(), "owner must be unchanged after a rejected call")

	// The owner can transfer ownership.
	testing.SetRealm(testing.NewUserRealm(alice))
	ChangeOwner(cross(cur), bob)
	uassert.Equal(t, Owner().String(), bob.String(), "owner must change after the current owner calls ChangeOwner")

	// The former owner has lost authority.
	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.AbortsWithMessage(t, cur, "caller isn't the owner", func() {
		ChangeOwner(cross(cur), alice)
	}, "the former owner must lose authority after transfer")
}
