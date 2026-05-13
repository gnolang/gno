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

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, o.TransferOwnership(0, cur, bob))
	}(cross)

	uassert.Equal(t, o.Owner(), bob)
}

func TestTransferOwnershipUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorContains(t, o.TransferOwnership(0, cur, bob), ErrUnauthorized.Error())
	}(cross)
}

func TestDropOwnership(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, o.DropOwnership(0, cur), "DropOwnership failed")
	}(cross)

	uassert.Empty(t, o.Owner(), "owner should be empty")
}

func TestDropOwnershipUnauthorized(t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorContains(t, o.DropOwnership(0, cur), ErrUnauthorized.Error())
	}(cross)
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

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		uassert.ErrorContains(t, o.TransferOwnership(0, cur, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t,
			o.TransferOwnership(0, cur, "10000000001000000000100000000010000000001000000000"),
			ErrInvalidAddress.Error())
	}(cross)
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
