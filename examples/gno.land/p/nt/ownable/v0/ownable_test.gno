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

func TestNewWithAddress(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	got := o.Owner()
	uassert.Equal(t, got, alice)
}

func TestTransferOwnership(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewUserRealm(alice))
	o := NewWithAddress(alice)
	func(cur realm) {
		err := o.TransferOwnership(0, cur, bob)
		urequire.NoError(t, err)
	}(cross(cur))

	got := o.Owner()
	uassert.Equal(t, got, bob)
}

func TestTransferOwnershipUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorContains(t, o.TransferOwnership(0, cur, bob), ErrUnauthorized.Error())
	}(cross(cur))
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

func TestDropOwnershipUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorContains(t, o.DropOwnership(0, cur), ErrUnauthorized.Error())
	}(cross(cur))
}

func TestOwnedBy(cur realm, t *testing.T) {
	o := NewWithAddress(alice)
	uassert.True(t, o.OwnedBy(alice))
	uassert.False(t, o.OwnedBy(bob))
}

func TestAssertOwnedBy(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	// Should not panic.
	o.AssertOwnedBy(alice)

	uassert.PanicsWithMessage(t, cur, ErrUnauthorized.Error(), func() {
		o.AssertOwnedBy(bob)
	})
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

func TestNilReceiver(cur realm, t *testing.T) {
	var o *Ownable

	owner := o.Owner()
	if owner != address("") {
		t.Errorf("expected empty address but got %v", owner)
	}

	isOwner := o.OwnedBy(alice)
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
	o.AssertOwnedBy(alice)
}
