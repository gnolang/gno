package ownable

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

var (
	alice = testutils.TestAddress("alice")
	bob   = testutils.TestAddress("bob")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // TODO(bug): should not be needed

	o := NewOwnable()
	got := o.Owner()
	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestNewWithAddress(t *testing.T) {
	o := NewOwnableWithAddress(alice)

	got := o.Owner()
	if alice != got {
		t.Fatalf("Expected %s, got: %s", alice, got)
	}
}

func TestOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := NewOwnable()
	expected := alice
	got := o.Owner()
	if expected != got {
		t.Fatalf("Expected %s, got: %s", expected, got)
	}
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := NewOwnable()

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

	o := NewOwnable()
	unauthorizedCaller := bob

	std.TestSetRealm(std.NewUserRealm(unauthorizedCaller))
	std.TestSetOrigCaller(unauthorizedCaller) // TODO(bug): should not be needed

	err := o.CallerIsOwner()
	if err == nil {
		t.Fatalf("Expected %s to not be owner", unauthorizedCaller)
	}
}

func TestDropOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := NewOwnable()

	err := o.DropOwnership()
	if err != nil {
		t.Fatalf("DropOwnership failed, %v", err)
	}

	owner := o.Owner()
	if owner != "" {
		t.Fatalf("Expected owner to be empty, not %s", owner)
	}
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))
	std.TestSetOrigCaller(alice) // TODO(bug): should not be needed

	o := NewOwnable()

	std.TestSetRealm(std.NewUserRealm(bob))
	std.TestSetOrigCaller(bob) // TODO(bug): should not be needed

	err := o.TransferOwnership(alice)
	if err != ErrUnauthorized {
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}

	err = o.DropOwnership()
	if err != ErrUnauthorized {
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}
}

func TestErrInvalidAddress(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(alice))

	o := NewOwnable()

	err := o.TransferOwnership("")
	if err != ErrInvalidAddress {
		t.Fatalf("Should've been ErrInvalidAddress, was %v", err)
	}

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	if err != ErrInvalidAddress {
		t.Fatalf("Should've been ErrInvalidAddress, was %v", err)
	}
}
