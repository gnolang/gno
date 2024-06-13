package ownable

import (
	"std"
	"testing"
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
	if firstCaller != got {
		t.Fatalf("Expected %s, got: %s", firstCaller, got)
	}
}

func TestOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()
	expected := firstCaller
	got := o.Owner()
	if expected != got {
		t.Fatalf("Expected %s, got: %s", expected, got)
	}
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

	err := o.TransferOwnership(secondCaller)
	if err != nil {
		t.Fatalf("TransferOwnership failed, %v", err)
	}

	got := o.Owner()
	if secondCaller != got {
		t.Fatalf("Expected: %s, got: %s", secondCaller, got)
	}
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

	unauthorizedCaller := secondCaller

	std.TestSetRealm(std.NewUserRealm(unauthorizedCaller))
	std.TestSetOrigCaller(unauthorizedCaller) // TODO(bug): should not be needed

	err := o.CallerIsOwner()
	if err == nil {
		t.Fatalf("Expected %s to not be owner", unauthorizedCaller)
	}
}

func TestDropOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

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
	std.TestSetRealm(std.NewUserRealm(firstCaller))
	std.TestSetOrigCaller(firstCaller) // TODO(bug): should not be needed

	o := New()

	std.TestSetRealm(std.NewUserRealm(secondCaller))
	std.TestSetOrigCaller(secondCaller) // TODO(bug): should not be needed

	err := o.TransferOwnership(firstCaller)
	if err != ErrUnauthorized {
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}

	err = o.DropOwnership()
	if err != ErrUnauthorized {
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}
}

func TestErrInvalidAddress(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstCaller))

	o := New()

	err := o.TransferOwnership("")
	if err != ErrInvalidAddress {
		t.Fatalf("Should've been ErrInvalidAddress, was %v", err)
	}

	err = o.TransferOwnership("10000000001000000000100000000010000000001000000000")
	if err != ErrInvalidAddress {
		t.Fatalf("Should've been ErrInvalidAddress, was %v", err)
	}
}
