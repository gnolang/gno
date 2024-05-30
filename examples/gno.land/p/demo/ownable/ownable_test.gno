package ownable

import (
	"std"
	"testing"
)

var (
	firstAddr  = std.Address("g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")
	secondAddr = std.Address("g1r0mlnkc05z0fv49km99z60qnp95tengyqfdr02")
)

func TestNew(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstAddr))

	result := New()
	if firstAddr != result.owner {
		t.Fatalf("Expected %s, got: %s\n", firstAddr, result.owner)
	}
}

func TestOwner(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstAddr))

	result := New()
	resultOwner := result.Owner()

	expected := firstAddr
	if resultOwner != expected {
		t.Fatalf("Expected %s, got: %s\n", expected, result)
	}
}

func TestTransferOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstAddr))
	o := New()

	err := o.TransferOwnership(secondAddr)
	if err != nil {
		t.Fatalf("TransferOwnership failed, %v", err)
	}

	result := o.Owner()
	if secondAddr != result {
		t.Fatalf("Expected: %s, got: %s\n", secondAddr, result)
	}
}

func TestCallerIsOwner(t *testing.T) {
	std.TestSetOrigCaller(firstAddr)

	o := New()

	unauthorizedCaller := secondAddr

	std.TestSetOrigCaller(unauthorizedCaller)

	if prevRealmIsOwner(o.owner) {
		t.Fatalf("Expected %s to not be owner\n", unauthorizedCaller)
	}
}

func TestDropOwnership(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstAddr))

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
	std.TestSetOrigCaller(firstAddr)

	o := New()

	std.TestSetOrigCaller(secondAddr)

	err := o.TransferOwnership(firstAddr)
	if err != ErrUnauthorized {
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}

	err = o.DropOwnership()
	if err != ErrUnauthorized {
		t.Fatalf("Should've been ErrUnauthorized, was %v", err)
	}
}

func TestErrInvalidAddress(t *testing.T) {
	std.TestSetRealm(std.NewUserRealm(firstAddr))

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
