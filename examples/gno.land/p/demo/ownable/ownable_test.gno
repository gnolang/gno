package ownable

import (
	"std"
	"testing"

	"gno.land/p/demo/helper"
)

var (
	firstCaller  = std.Address("g1l9aypkr8xfvs82zeux486ddzec88ty69lue9de")
	secondCaller = std.Address("g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa")
)

func TestNew(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)

	result := New()
	if firstCaller != result.owner {
		t.Fatalf("Expected %s, got: %s\n", firstCaller, result.owner)
	}
}

func TestOwner(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)

	result := New()
	resultOwner := result.Owner()

	expected := firstCaller
	if resultOwner != expected {
		t.Fatalf("Expected %s, got: %s\n", expected, result)
	}
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)
	o := New()

	err := o.TransferOwnership(secondCaller)
	if err != nil {
		t.Fatalf("TransferOwnership failed, %v", err)
	}

	result := o.Owner()
	if secondCaller != result {
		t.Fatalf("Expected: %s, got: %s\n", secondCaller, result)
	}
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)

	o := New()
	unauthorizedCaller := secondCaller

	std.TestSetOrigCaller(unauthorizedCaller)

	if helper.PrevRealmIsOwner(o.owner) {
		t.Fatalf("Expected %s to not be owner\n", unauthorizedCaller)
	}
}

func TestDropOwnership(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)

	o := New()

	err := o.DropOwnership()
	if err != nil {
		t.Fatalf("DropOwnership failed, %v", err)
	}

	owner := o.Owner()
	if owner != "" {
		t.Fatalf("Expected owner to be empty, not %s\n", owner)
	}
}

// Errors

func TestErrUnauthorized(t *testing.T) {
	std.TestSetOrigCaller(firstCaller)

	o := New()

	std.TestSetOrigCaller(secondCaller)

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
	std.TestSetOrigCaller(firstCaller)

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
