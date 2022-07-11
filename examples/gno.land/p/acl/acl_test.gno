package acl

import (
	"std"
	"testing"

	"gno.land/p/testutils"
)

func Test(t *testing.T) {
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")

	shouldNotHasRole(t, test1, "foo")
	shouldNotHasRole(t, test2, "foo")
}

func shouldHasRole(t *testing.T, addr std.Address, role string) {
	t.Helper()
	check := HasRole(addr, role)
	if !check {
		t.Errorf("%q should has role %q", addr.String(), role)
	}
}

func shouldNotHasRole(t *testing.T, addr std.Address, role string) {
	t.Helper()
	check := HasRole(addr, role)
	if check {
		t.Errorf("%q should not has role %q", addr.String(), role)
	}
}
