package ownable2step

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestNew(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()
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

func TestInitiateTransferOwnership(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)
}

func TestTransferOwnership(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)

	t.SetPreviousRealm(std.NewUserRealm(bob))

	err = o.AcceptOwnership()
	urequire.NoError(t, err)

	owner = o.Owner()
	pendingOwner = o.PendingOwner()

	uassert.Equal(t, owner, bob)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestCallerIsOwner(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()
	unauthorizedCaller := bob

	t.SetPreviousRealm(std.NewUserRealm(unauthorizedCaller))

	uassert.False(t, o.CallerIsOwner())
}

func TestDropOwnership(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()

	err := o.DropOwnership()
	urequire.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()

	t.SetPreviousRealm(std.NewUserRealm(bob))

	uassert.ErrorContains(t, o.TransferOwnership(alice), ErrUnauthorized.Error())
	uassert.ErrorContains(t, o.DropOwnership(), ErrUnauthorized.Error())
}

func TestErrInvalidAddress(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestErrNoPendingOwner(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()

	err := o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

func TestErrPendingUnauthorized(t *testing.T) {
	t.SetPreviousRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	t.SetPreviousRealm(std.NewUserRealm(alice))

	err = o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
}
