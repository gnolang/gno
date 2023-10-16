package authctx

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestNewOrigAuthContext(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewOrigAuthContext()

	assertSameAddresses(t, test1, auth.Addr())
	assertSameStrings(t, "orig:"+string(test1), auth.String())
	if !auth.HasPerm("foo") {
		t.Errorf("should have all perms")
	}
}

func TestNewPrevAuthContext(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewPrevAuthContext()

	assertSameAddresses(t, test1, auth.Addr())
	assertSameStrings(t, "prev:"+string(test1), auth.String())
	if !auth.HasPerm("foo") {
		t.Errorf("should have all perms")
	}
}

// TODO: TestDelegations

func assertSameStrings(t *testing.T, expected, got string) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}

func assertSameAddresses(t *testing.T, expected, got std.Address) {
	t.Helper()
	if expected != got {
		t.Errorf("expected %q, got %q.", expected, got)
	}
}
