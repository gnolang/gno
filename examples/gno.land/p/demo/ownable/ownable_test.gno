package ownable

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
	testing.SetOriginCaller(alice)

	o := New()
	got := o.Owner()
	uassert.Equal(t, got, alice)
}

func TestNewWithAddress(t *testing.T) {
	o := NewWithAddress(alice)

	got := o.Owner()
	uassert.Equal(t, got, alice)
}

func TestTransferOwnership(t *testing.T) {
	testing.SetOriginCaller(alice)

	o := New()

	err := o.TransferOwnership(bob)
	urequire.NoError(t, err)

	got := o.Owner()

	uassert.Equal(t, got, bob)
}

func TestOwnedByPrevious(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(alice))

	o := New()
	unauthorizedCaller := bob

	testing.SetOriginCaller(unauthorizedCaller)

	uassert.False(t, o.OwnedByPrevious())
}

func TestDropOwnership(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.DropOwnership()
	urequire.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	uassert.Empty(t, owner, "owner should be empty")
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

func TestAssertOwnedByPrevious(t *testing.T) {
	testing.SetRealm(std.NewUserRealm(alice))
	testing.SetOriginCaller(alice)

	o := New()

	// Should not panic when caller is owner
	o.AssertOwnedByPrevious()

	// Should panic when caller is not owner
	testing.SetRealm(std.NewUserRealm(bob))
	testing.SetOriginCaller(bob)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic but got none")
		}
		if r != ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized but got %v", r)
		}
	}()
	o.AssertOwnedByPrevious()
}

func TestNilReceiver(t *testing.T) {
	var o *Ownable

	owner := o.Owner()
	if owner != std.Address("") {
		t.Errorf("expected empty address but got %v", owner)
	}

	isOwner := o.OwnedByPrevious()
	uassert.False(t, isOwner)

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic but got none")
		}
		if r != ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized but got %v", r)
		}
	}()
	o.AssertOwnedByPrevious()
}
