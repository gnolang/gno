package ownable

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/urequire"
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

// embed ownable and test
// test safeownable

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // TODO(bug): should not be needed

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
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership(bob)

	got := o.Owner()
	uassert.Equal(t, got, bob)
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()
	unauthorizedCaller := bob

	std.TestSetRealm(std.NewUserRealm(unauthorizedCaller))
	std.TestSetOrigCaller(unauthorizedCaller) // TODO(bug): should not be needed

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
	std.TestSetOrigCaller(alice) // TODO(bug): should not be needed

	o := New()

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob) // TODO(bug): should not be needed

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
