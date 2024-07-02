package ownable

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	firstCaller  = testutils.TestAddress("first")
	secondCaller = testutils.TestAddress("second")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	o := New()
	got := o.Owner()
	uassert.Equal(t, firstCaller, got)
}

func TestNewWithAddress(t *testing.T) {
	o := NewWithAddress(firstCaller)

	got := o.Owner()
	uassert.Equal(t, firstCaller, got)
}

func TestOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()
	expected := firstCaller
	got := o.Owner()
	uassert.Equal(t, expected, got)
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

	err := o.TransferOwnership(secondCaller)
	uassert.NoError(t, err, "TransferOwnership failed")

	got := o.Owner()
	uassert.Equal(t, secondCaller, got)
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()
	unauthorizedCaller := secondCaller

	std.TestSetRealm(std.NewUserRealm(unauthorizedCaller))
	std.TestSetOrigCaller(unauthorizedCaller) // TODO(bug): should not be needed

	err := o.CallerIsOwner()
	uassert.Error(t, err) // XXX: IsError(..., unauthorizedCaller)
}

func TestDropOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

	err := o.DropOwnership()
	uassert.NoError(t, err, "DropOwnership failed")

	owner := o.Owner()
	if owner != "" { // XXX: uassert.Empty...
		t.Fatalf("Expected owner to be empty, not %s", owner)
	}
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	o := New()

	std.TestSetRealm(std.NewUserRealm(secondCaller))
	std.TestSetOrigCaller(secondCaller) // TODO(bug): should not be needed

	err := o.TransferOwnership(firstCaller)
	uassert.ErrorContains(t, err, ErrUnauthorized.Error())

	err = o.DropOwnership()
	uassert.ErrorContains(t, err, ErrUnauthorized.Error())
}

func TestErrInvalidAddress(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

	err := o.TransferOwnership("")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	uassert.ErrorContains(t, err, ErrInvalidAddress.Error())
}
