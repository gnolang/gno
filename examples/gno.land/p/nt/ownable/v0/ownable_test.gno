package ownable

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)


// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm
var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestNewWithAddress(cur realm, t *testing.T) {
	o := NewWithAddress(alice)
	uassert.Equal(t, o.Owner(), alice)
}

func TestTransferOwnership(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, o.TransferOwnership(0, cur, bob))
	}(cross2(cur))

	uassert.Equal(t, o.Owner(), bob)
}

func TestTransferOwnershipUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorContains(t, o.TransferOwnership(0, cur, bob), ErrUnauthorized.Error())
	}(cross2(cur))
}

func TestDropOwnership(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		urequire.NoError(t, o.DropOwnership(0, cur), "DropOwnership failed")
	}(cross2(cur))

	uassert.Empty(t, o.Owner(), "owner should be empty")
}

func TestDropOwnershipUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	func(cur realm) {
		uassert.ErrorContains(t, o.DropOwnership(0, cur), ErrUnauthorized.Error())
	}(cross2(cur))
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
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	func(cur realm) {
		uassert.ErrorContains(t, o.TransferOwnership(0, cur, ""), ErrInvalidAddress.Error())
		uassert.ErrorContains(t,
			o.TransferOwnership(0, cur, "10000000001000000000100000000010000000001000000000"),
			ErrInvalidAddress.Error())
	}(cross2(cur))
}

func TestNilReceiver(cur realm, t *testing.T) {
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
