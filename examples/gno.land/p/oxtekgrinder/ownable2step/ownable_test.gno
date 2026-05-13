package ownable2step

import (
	"chain"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestNew(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)
	got := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, got, alice)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestNewWithAddress(t *testing.T) {
	o := NewWithAddress(alice)

	got := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, got, alice)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestInitiateTransferOwnership(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	err := o.TransferOwnership(0, cur, bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)
}

func TestTransferOwnership(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	err := o.TransferOwnership(0, cur, bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)

	testing.SetRealm(testing.NewUserRealm(bob))

	err = o.AcceptOwnership(0, cur)
	urequire.NoError(t, err)

	owner = o.Owner()
	pendingOwner = o.PendingOwner()

	uassert.Equal(t, owner, bob)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestOwnedByCurrent(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.False(t, o.OwnedByCurrent(0, cur))
}

func TestOwnedByCurrentUnauthorized(cur realm, t *testing.T) {
	// Mirrors the original: NewWithOrigin called from inside a code-realm
	// scope where the EOA origin is alice, then OwnedByCurrent checked
	// from outside that scope (alice still current). Final check: when
	// bob becomes current, OwnedByCurrent returns false.
	const testPkgPath = "gno.land/r/test/test"
	aliceRealm := testing.MakeRealm(alice, "", testing.OriginRealm())
	crossed := testing.MakeRealm(chain.PackageAddress(testPkgPath), testPkgPath, aliceRealm)
	o := NewWithOrigin(0, crossed)

	uassert.True(t, o.OwnedByCurrent(0, aliceRealm))

	bobRealm := testing.MakeRealm(bob, "", testing.OriginRealm())
	uassert.False(t, o.OwnedByCurrent(0, bobRealm))
}

func TestOwnedByPrevious(cur realm, t *testing.T) {
	// Mirrors the original: owner is alice (via SetOriginCaller(alice));
	// a code realm has crossed into us so cur.Previous() == alice; the
	// OwnedByPrevious check should authenticate alice as caller.
	const testPkgPath = "gno.land/r/test/test"
	aliceRealm := testing.MakeRealm(alice, "", testing.OriginRealm())
	o := New(0, aliceRealm)

	crossed := testing.MakeRealm(chain.PackageAddress(testPkgPath), testPkgPath, aliceRealm)
	uassert.True(t, o.OwnedByPrevious(0, crossed))
}

func TestOwnedByPreviousUnauthorized(cur realm, t *testing.T) {
	// Mirrors the original: owner is alice; then bob (unauthorized) is
	// the current realm; a code realm has crossed into bob, so
	// cur.Previous() == bob — not the owner. OwnedByPrevious returns false.
	const testPkgPath = "gno.land/r/test/test"
	aliceRealm := testing.MakeRealm(alice, "", testing.OriginRealm())
	o := New(0, aliceRealm)

	bobRealm := testing.MakeRealm(bob, "", testing.OriginRealm())
	crossed := testing.MakeRealm(chain.PackageAddress(testPkgPath), testPkgPath, bobRealm)
	uassert.False(t, o.OwnedByPrevious(0, crossed))
}

func TestDropOwnership(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	err := o.DropOwnership(0, cur)
	urequire.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

func TestDropOwnershipVulnerability(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := New(0, cur)

	// alice initiates ownership transfer to bob
	err := o.TransferOwnership(0, cur, bob)
	urequire.NoError(t, err)

	// alice drops ownership while ownership transfer is pending
	err = o.DropOwnership(0, cur)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()
	uassert.Empty(t, owner, "owner should be empty")
	uassert.Empty(t, pendingOwner.String(), "pending owner should be empty")

	// verify bob can no longer claim ownership
	testing.SetRealm(testing.NewUserRealm(bob))
	err = o.AcceptOwnership(0, cur)
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

// Errors

func TestErrUnauthorized(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	testing.SetRealm(testing.NewUserRealm(bob))

	uassert.ErrorContains(t, o.TransferOwnership(0, cur, alice), ErrUnauthorized.Error())
	uassert.ErrorContains(t, o.DropOwnership(0, cur), ErrUnauthorized.Error())
}

func TestErrInvalidAddress(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	err := o.TransferOwnership(0, cur, "")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership(0, cur, "10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestErrNoPendingOwner(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	err := o.AcceptOwnership(0, cur)
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

func TestErrPendingUnauthorized(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New(0, cur)

	err := o.TransferOwnership(0, cur, bob)
	urequire.NoError(t, err)

	testing.SetRealm(testing.NewUserRealm(alice))

	err = o.AcceptOwnership(0, cur)
	uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
}
