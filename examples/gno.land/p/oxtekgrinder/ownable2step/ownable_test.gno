package ownable2step

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestNewWithAddress(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	got := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, got, alice)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestInitiateTransferOwnership(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		err := o.TransferOwnership(0, cur, bob)
		urequire.NoError(t, err)
	}(cross(cur))

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)
}

func TestTransferOwnership(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		err := o.TransferOwnership(0, cur, bob)
		urequire.NoError(t, err)
	}(cross(cur))

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		err := o.AcceptOwnership(0, cur)
		urequire.NoError(t, err)
	}(cross(cur))

	owner = o.Owner()
	pendingOwner = o.PendingOwner()

	uassert.Equal(t, owner, bob)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestOwnedBy(cur realm, t *testing.T) {
	o := NewWithAddress(alice)
	uassert.True(t, o.OwnedBy(alice))
	uassert.False(t, o.OwnedBy(bob))
}

func TestDropOwnership(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		err := o.DropOwnership(0, cur)
		urequire.NoError(t, err, "DropOwnership failed")
	}(cross(cur))

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

func TestDropOwnershipVulnerability(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		// alice initiates ownership transfer to bob
		err := o.TransferOwnership(0, cur, bob)
		urequire.NoError(t, err)

		// alice drops ownership while ownership transfer is pending
		err = o.DropOwnership(0, cur)
		urequire.NoError(t, err)
	}(cross(cur))

	owner := o.Owner()
	pendingOwner := o.PendingOwner()
	uassert.Empty(t, owner, "owner should be empty")
	uassert.Empty(t, pendingOwner.String(), "pending owner should be empty")

	// verify bob can no longer claim ownership
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		err := o.AcceptOwnership(0, cur)
		uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
	}(cross(cur))
}

// Errors

func TestErrUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorContains(t, o.TransferOwnership(0, cur, alice), ErrUnauthorized.Error())
		uassert.ErrorContains(t, o.DropOwnership(0, cur), ErrUnauthorized.Error())
	}(cross(cur))
}

func TestErrInvalidAddress(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		err := o.TransferOwnership(0, cur, "")
		uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

		err = o.TransferOwnership(0, cur, "10000000001000000000100000000010000000001000000000")
		uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
	}(cross(cur))
}

func TestErrNoPendingOwner(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		err := o.AcceptOwnership(0, cur)
		uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
	}(cross(cur))
}

func TestErrPendingUnauthorized(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		err := o.TransferOwnership(0, cur, bob)
		urequire.NoError(t, err)

		// alice (still the owner, NOT the pending) tries to AcceptOwnership.
		err = o.AcceptOwnership(0, cur)
		uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
	}(cross(cur))
}
