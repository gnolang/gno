package ownable

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // TODO(bug): should not be needed

	o := New()
	got := o.Owner()
	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestNewWithAddress(t *testing.T) {
	o := NewWithAddress(alice)

	got := o.Owner()
	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()
	expected := alice
	got := o.Owner()
	uassert.Equal(t, expected, got)
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership(bob)
	if err != nil {
		t.Fatalf("TransferOwnership failed, %v", err)
	}

	got := o.Owner()
	if bob != got {
		t.Fatalf("Expected: %s, got: %s", bob, got)
	}
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()
	unauthorizedCaller := bob

	std.TestSetRealm(std.NewUserRealm(unauthorizedCaller))
	std.TestSetOrigCaller(unauthorizedCaller) // TODO(bug): should not be needed

	err := o.CallerIsOwner()
	uassert.Error(t, err) // XXX: IsError(..., unauthorizedCaller)
}

func TestDropOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.DropOwnership()
	uassert.NoError(t, err, "DropOwnership failed")

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

	err := o.TransferOwnership(alice)
	if err != ErrUnauthorized {
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}

	err = o.DropOwnership()
	uassert.ErrorContains(t, err, ErrUnauthorized.Error())
}

func TestErrInvalidAddress(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := New()

	err := o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}
