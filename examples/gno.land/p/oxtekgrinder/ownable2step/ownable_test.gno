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

func TestNew(t *testing.T) {
	testing.SetOriginCaller(alice)

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
	testing.SetOriginCaller(alice)

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)
}

func TestTransferOwnership(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()

	uassert.Equal(t, owner, alice)
	uassert.Equal(t, pendingOwner, bob)

	testing.SetOriginCaller(bob)

	err = o.AcceptOwnership()
	urequire.NoError(t, err)

	owner = o.Owner()
	pendingOwner = o.PendingOwner()

	uassert.Equal(t, owner, bob)
	uassert.Equal(t, pendingOwner.String(), "")
}

func TestOwnedByCurrent(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()
	unauthorizedCaller := bob
	testing.SetOriginCaller(unauthorizedCaller)
	uassert.False(t, o.OwnedByCurrent())
}

func TestOwnedByCurrentUnauthorized(t *testing.T) {
	testing.SetOriginCaller(alice)
	testing.SetRealm(testing.NewUserRealm(alice))

	var o *Ownable2Step
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
		o = NewWithOrigin()
	}()

	uassert.True(t, o.OwnedByCurrent())

	unauthorizedCaller := bob
	testing.SetRealm(testing.NewUserRealm(unauthorizedCaller))
	uassert.False(t, o.OwnedByCurrent())
}

func TestOwnedByPrevious(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
		uassert.True(t, o.OwnedByPrevious())
	}()
}

func TestOwnedByPreviousUnauthorized(t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))

	o := New()
	unauthorizedCaller := bob
	testing.SetRealm(testing.NewUserRealm(unauthorizedCaller))
	func() {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
		uassert.False(t, o.OwnedByPrevious())
	}()
}

func TestDropOwnership(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()

	err := o.DropOwnership()
	urequire.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
}

func TestDropOwnershipVulnerability(t *testing.T) {
	testing.SetOriginCaller(alice)
	o := New()

	// alice initiates ownership transfer to bob
	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	// alice drops ownership while ownership transfer is pending
	err = o.DropOwnership()
	urequire.NoError(t, err)

	owner := o.Owner()
	pendingOwner := o.PendingOwner()
	uassert.Empty(t, owner, "owner should be empty")
	uassert.Empty(t, pendingOwner.String(), "pending owner should be empty")

	// verify bob can no longer claim ownership
	testing.SetOriginCaller(bob)
	err = o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()

	testing.SetOriginCaller(bob)

	uassert.ErrorContains(t, o.TransferOwnership(alice), ErrUnauthorized.Error())
	uassert.ErrorContains(t, o.DropOwnership(), ErrUnauthorized.Error())
}

func TestErrInvalidAddress(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()

	err := o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestErrNoPendingOwner(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()

	err := o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrNoPendingOwner.Error())
}

func TestErrPendingUnauthorized(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	testing.SetOriginCaller(alice)

	err = o.AcceptOwnership()
	uassert.ErrorContains(t, err, ErrPendingUnauthorized.Error())
}
