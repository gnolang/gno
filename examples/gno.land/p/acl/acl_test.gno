package acl

import (
	"std"
	"testing"

	"gno.land/p/testutils"
)

func Test(t *testing.T) {
	test1 := testutils.TestAddress("test1")
	test2 := testutils.TestAddress("test2")

	dir := New()

	shouldNotHasRole(t, dir, test1, "foo")
	shouldNotHasRole(t, dir, test2, "foo")
	shouldNotHasPerm(t, dir, test1, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, test2, "write", "r/boards:gnolang/1")
}

func shouldHasRole(t *testing.T, dir *Directory, addr std.Address, role string) {
	t.Helper()
	check := dir.HasRole(addr, role)
	if !check {
		t.Errorf("%q should has role %q", addr.String(), role)
	}
}

func shouldNotHasRole(t *testing.T, dir *Directory, addr std.Address, role string) {
	t.Helper()
	check := dir.HasRole(addr, role)
	if check {
		t.Errorf("%q should not has role %q", addr.String(), role)
	}
}

func shouldHasPerm(t *testing.T, dir *Directory, addr std.Address, verb string, resource string) {
	t.Helper()
	check := dir.HasPerm(addr, verb, resource)
	if !check {
		t.Errorf("%q should has perm for %q - %q", addr.String(), verb, resource string)
	}
}

func shouldNotHasPerm(t *testing.T, dir *Directory, addr std.Address, verb string, resource string) {
	t.Helper()
	check := dir.HasPerm(addr, verb, resource)
	if check {
		t.Errorf("%q should not has perm for %q - %q", addr.String(), verb, resource)
	}
}
