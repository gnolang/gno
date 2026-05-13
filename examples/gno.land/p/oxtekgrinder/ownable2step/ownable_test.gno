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

func TestOwnedByPrevious(cur realm, t *testing.T) {
	// Construct an Ownable2Step whose owner is the realm we expect
	// cur.Previous() to surface when SetRealm overrides the test's cur
	// to a CodeRealm — that's the test frame's own package address
	// (set by X_setContext's CodeRealm-override branch).
	pkgPath := "gno.land/p/oxtekgrinder/ownable2step"
	o := NewWithAddress(chain.PackageAddress(pkgPath))

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
	uassert.True(t, o.OwnedByPrevious(0, cur))
}

func TestOwnedByPreviousUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(bob)

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
	uassert.False(t, o.OwnedByPrevious(0, cur))
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
