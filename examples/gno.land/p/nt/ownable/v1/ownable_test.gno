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

// Tests use plain testing.SetRealm without cross() scaffolding —
// the v3a-aligned override walk handles the user-realm override
// frame as the EOA boundary so runtime.Caller() resolves correctly.

func TestNewWithAddress(cur realm, t *testing.T) {
	o := NewWithAddress(alice)
	uassert.Equal(t, o.Owner(), alice)
}

func TestOwnedBy(cur realm, t *testing.T) {
	o := NewWithAddress(alice)
	uassert.True(t, o.OwnedBy(alice))
	uassert.False(t, o.OwnedBy(bob))
	uassert.False(t, o.OwnedBy(""))
}

func TestOwnedByNilReceiver(cur realm, t *testing.T) {
	var o *Ownable
	uassert.False(t, o.OwnedBy(alice))
}

func TestAssertOwnedBy(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	// alice is owner: AssertOwnedBy should not panic.
	o.AssertOwnedBy(alice)

	// bob is not owner: AssertOwnedBy should panic.
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		o.AssertOwnedBy(bob)
	}()
	uassert.True(t, panicked, "expected panic for non-owner")
}

func TestOwnerNilReceiver(cur realm, t *testing.T) {
	var o *Ownable
	uassert.Equal(t, o.Owner(), address(""))
}

func TestTransferOwnership(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	urequire.NoError(t, o.TransferOwnership(bob))
	uassert.Equal(t, o.Owner(), bob)
}

func TestTransferOwnershipUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.ErrorContains(t, o.TransferOwnership(bob), ErrUnauthorized.Error())
}

func TestTransferOwnershipInvalidAddress(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	uassert.ErrorContains(t, o.TransferOwnership(""), ErrInvalidAddress.Error())
}

func TestDropOwnership(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(alice))
	urequire.NoError(t, o.DropOwnership())
	uassert.Empty(t, string(o.Owner()), "owner should be empty after drop")
}

func TestDropOwnershipUnauthorized(cur realm, t *testing.T) {
	o := NewWithAddress(alice)

	testing.SetRealm(testing.NewUserRealm(bob))
	uassert.ErrorContains(t, o.DropOwnership(), ErrUnauthorized.Error())
}
