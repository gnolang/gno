package ownable

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

var (
	firstCaller  = std.Address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	secondCaller = std.Address("g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa")
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
	if err != ErrUnauthorized { // XXX: uassert.IsError
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}

	err = o.DropOwnership()
	if err != ErrUnauthorized { // XXX: uassert.IsError
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}
}

func TestErrInvalidAddress(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

	err := o.TransferOwnership("")
	if err != ErrInvalidAddress { // XXX: uassert.IsError
		t.Fatalf("Should've been ErrInvalidAddress, was %v", err)
	}

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	if err != ErrInvalidAddress { // XXX: uassert.IsError
		t.Fatalf("Should've been ErrInvalidAddress, was %v", err)
	}
}
