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

func TestNewWithAddress(t *testing.T) {
	o := NewWithAddress(alice)

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner().String(), "")
}

func TestInitiateTransferOwnership(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = o.TransferOwnership(0, cur, bob) }(cross)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner(), bob)
}

func TestTransferOwnership(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = o.TransferOwnership(0, cur, bob) }(cross)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner(), bob)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) { err = o.AcceptOwnership(0, cur) }(cross)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), bob)
	uassert.Equal(t, o.PendingOwner().String(), "")
}

func TestOwnedBy(t *testing.T) {
	o := NewWithAddress(alice)
	uassert.True(t, o.OwnedBy(alice))
	uassert.False(t, o.OwnedBy(bob))
}

func TestDropOwnership(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = o.DropOwnership(0, cur) }(cross)
	urequire.NoError(t, err, "DropOwnership failed")

	uassert.Empty(t, o.Owner(), "owner should be empty")
}

func TestDropOwnershipVulnerability(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	// alice initiates ownership transfer to bob
	func(cur realm) { err = o.TransferOwnership(0, cur, bob) }(cross)
	urequire.NoError(t, err)

	// alice drops ownership while ownership transfer is pending
	func(cur realm) { err = o.DropOwnership(0, cur) }(cross)
	urequire.NoError(t, err)

	uassert.Empty(t, o.Owner(), "owner should be empty")
	uassert.Empty(t, o.PendingOwner().String(), "pending owner should be empty")

	// verify bob can no longer claim ownership
	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) { err = o.AcceptOwnership(0, cur) }(cross)
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	var err error
	func(cur realm) { err = o.TransferOwnership(0, cur, alice) }(cross)
	uassert.ErrorContains(t, err, ErrUnauthorized.Error())
	func(cur realm) { err = o.DropOwnership(0, cur) }(cross)
	uassert.ErrorContains(t, err, ErrUnauthorized.Error())
}

func TestErrInvalidAddress(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = o.TransferOwnership(0, cur, "") }(cross)
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	func(cur realm) {
		err = o.TransferOwnership(0, cur, "10000000001000000000100000000010000000001000000000")
	}(cross)
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestErrNoPendingOwner(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = o.AcceptOwnership(0, cur) }(cross)
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

func TestErrPendingUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	func(cur realm) { err = o.TransferOwnership(0, cur, bob) }(cross)
	urequire.NoError(t, err)

	// alice (still the owner, NOT the pending) tries to AcceptOwnership.
	func(cur realm) { err = o.AcceptOwnership(0, cur) }(cross)
	uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
}
