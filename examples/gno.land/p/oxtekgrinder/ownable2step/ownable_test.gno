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

	err := o.TransferOwnership(alice, bob)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner(), bob)
}

func TestTransferOwnership(t *testing.T) {
	o := NewWithAddress(alice)

	err := o.TransferOwnership(alice, bob)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), alice)
	uassert.Equal(t, o.PendingOwner(), bob)

	err = o.AcceptOwnership(bob)
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

	err := o.DropOwnership(alice)
	urequire.NoError(t, err, "DropOwnership failed")

	uassert.Empty(t, o.Owner(), "owner should be empty")
}

func TestDropOwnershipVulnerability(t *testing.T) {
	o := NewWithAddress(alice)

	// alice initiates ownership transfer to bob
	err := o.TransferOwnership(alice, bob)
	urequire.NoError(t, err)

	// alice drops ownership while ownership transfer is pending
	err = o.DropOwnership(alice)
	urequire.NoError(t, err)

	uassert.Empty(t, o.Owner(), "owner should be empty")
	uassert.Empty(t, o.PendingOwner().String(), "pending owner should be empty")

	// verify bob can no longer claim ownership
	err = o.AcceptOwnership(bob)
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)

	uassert.ErrorContains(t, o.TransferOwnership(bob, alice), ErrUnauthorized.Error())
	uassert.ErrorContains(t, o.DropOwnership(bob), ErrUnauthorized.Error())
}

func TestErrInvalidAddress(t *testing.T) {
	o := NewWithAddress(alice)

	err := o.TransferOwnership(alice, "")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership(alice, "10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestErrNoPendingOwner(t *testing.T) {
	o := NewWithAddress(alice)

	err := o.AcceptOwnership(alice)
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

func TestErrPendingUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)

	err := o.TransferOwnership(alice, bob)
	urequire.NoError(t, err)

	err = o.AcceptOwnership(alice)
	uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
}
