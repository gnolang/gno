package auth

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestNewOrigAuth(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewOrigAuth()

	assertSameStrings(t, "orig", auth.Kind())
	assertSameAddresses(t, test1, auth.Addr())
	assertSameStrings(t, "orig:"+string(test1), auth.String())
	if !auth.HasPerm("foo") {
		t.Errorf("should have all perms")
	}
}

func TestNewPrevAuth(t *testing.T) {
	var test1 std.Address = testutils.TestAddress("test1")
	std.TestSetOrigCaller(test1)

	auth := NewPrevAuth()

	assertSameStrings(t, "prev", auth.Kind())
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
