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
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOriginCaller(alice)

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
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOriginCaller(alice)

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOriginCaller(alice)

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOriginCaller(bob)

	err = o.AcceptOwnership()
	urequire.NoError(t, err)

	owner = o.Owner()
	pendingOwner = o.PendingOwner()

	uassert.Equal(t, owner, bob)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOriginCaller(alice)

	o := New()
	unauthorizedCaller := bob

	std.TestSetRealm(std.NewUserRealm(unauthorizedCaller))
	std.TestSetOriginCaller(unauthorizedCaller)

	uassert.False(t, o.CallerIsOwner())
}

func TestDropOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.DropOwnership()
	urequire.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOriginCaller(alice)

	o := New()

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOriginCaller(bob)

	uassert.ErrorContains(t, o.TransferOwnership(alice), ErrUnauthorized.Error())
	uassert.ErrorContains(t, o.DropOwnership(), ErrUnauthorized.Error())
}

func TestErrInvalidAddress(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestErrNoPendingOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

func TestErrPendingUnauthorized(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	std.TestSetRealm(std.NewUserRealm(alice))

	err = o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
}
