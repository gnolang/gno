package ownable

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
}

func TestTransferOwnership(t *testing.T) {
	o := NewWithAddress(alice)

	err := o.TransferOwnership(alice, bob)
	urequire.NoError(t, err)

	uassert.Equal(t, o.Owner(), bob)
}

func TestTransferOwnershipUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)
	uassert.ErrorContains(t, o.TransferOwnership(bob, bob), ErrUnauthorized.Error())
}

func TestDropOwnership(t *testing.T) {
	o := NewWithAddress(alice)

	err := o.DropOwnership(alice)
	urequire.NoError(t, err, "DropOwnership failed")

	uassert.Empty(t, o.Owner(), "owner should be empty")
}

func TestDropOwnershipUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)
	uassert.ErrorContains(t, o.DropOwnership(bob), ErrUnauthorized.Error())
}

func TestOwnedBy(t *testing.T) {
	o := NewWithAddress(alice)
	uassert.True(t, o.OwnedBy(alice))
	uassert.False(t, o.OwnedBy(bob))
}

func TestAssertOwnedBy(t *testing.T) {
	o := NewWithAddress(alice)

	// Should not panic.
	o.AssertOwnedBy(alice)

	uassert.PanicsWithMessage(t, ErrUnauthorized.Error(), func() {
		o.AssertOwnedBy(bob)
	})
}

func TestErrInvalidAddress(t *testing.T) {
	o := NewWithAddress(alice)
	err := o.TransferOwnership(alice, "")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership(alice, "10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}

func TestNilReceiver(t *testing.T) {
	var o *Ownable

	owner := o.Owner()
	if owner != address("") {
		t.Errorf("expected empty address but got %v", owner)
	}

	uassert.False(t, o.OwnedBy(alice))

	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic but got none")
		}
		if r != ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized but got %v", r)
		}
	}()
	o.AssertOwnedBy(alice)
}
