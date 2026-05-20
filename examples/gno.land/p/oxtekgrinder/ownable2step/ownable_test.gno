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

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner().String(), "")
}

func TestInitiateTransferOwnership(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	err = o.TransferOwnership(bob)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner(), bob)
}

func TestTransferOwnership(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	err = o.TransferOwnership(bob)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner(), bob)

	testing.SetRealm(testing.NewUserRealm(bob))
	err = o.AcceptOwnership()
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), bob)
	uassert.Equal(t, o.PendingOwner().String(), "")
}

func TestOwnedBy(cur realm, t *testing.T) {
	o := NewWithAddress(alice)
	uassert.True(t, o.OwnedBy(alice))
	uassert.False(t, o.OwnedBy(bob))
}

func TestDropOwnership(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	err = o.DropOwnership()
	urequire.NoError(t, err, "DropOwnership failed")

	uassert.Empty(t, o.Owner(), "owner should be empty")
}

func TestDropOwnershipVulnerability(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	// alice initiates ownership transfer to bob
	err = o.TransferOwnership(bob)
	urequire.NoError(t, err)

	// alice drops ownership while ownership transfer is pending
	err = o.DropOwnership()
	urequire.NoError(t, err)

	uassert.Empty(t, o.Owner(), "owner should be empty")
	uassert.Empty(t, o.PendingOwner().String(), "pending owner should be empty")

	// verify bob can no longer claim ownership
	testing.SetRealm(testing.NewUserRealm(bob))
	err = o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

// Errors

func TestErrUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	var err error
	err = o.TransferOwnership(alice)
	uassert.ErrorContains(t, err, ErrUnauthorized.Error())
	err = o.DropOwnership()
	uassert.ErrorContains(t, err, ErrUnauthorized.Error())
}

func TestErrInvalidAddress(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	err = o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestErrNoPendingOwner(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	err = o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

func TestErrPendingUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	var err error
	err = o.TransferOwnership(bob)
	urequire.NoError(t, err)

	// alice (still the owner, NOT the pending) tries to AcceptOwnership.
	err = o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
}
