package acl

import (
	"std"
	"testing"

	"gno.land/p/testutils"
)

func Test(t *testing.T) {
	adm := testutils.TestAddress("admin")
	mod := testutils.TestAddress("mod")
	usr := testutils.TestAddress("user")
	cst := testutils.TestAddress("custom")

	dir := New()

	// by default, no one has perm.
	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, adm, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/boards:gnolang/1")

	// adding all the rights to admin.
	dir.AddUserPerm(adm, ".*", ".*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/boards:gnolang/1") // new
	shouldNotHasPerm(t, dir, mod, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/boards:gnolang/1") // new
	shouldNotHasPerm(t, dir, mod, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/boards:gnolang/1")

	// adding custom regexp rule for user "cst".
	dir.AddUserPerm(cst, "write", "r/boards:gnolang/.*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/boards:gnolang/1") // new
	shouldHasPerm(t, dir, adm, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/boards:gnolang/1")

	// adding a group perm for a new group.
	// no changes expected.
	dir.AddGroupPerm("mods", "role", "moderator")
	dir.AddGroupPerm("mods", "write", ".*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/boards:gnolang/1")

	// assigning the user "mod" to the "mods" group.
	dir.AddUserToGroup(mod, "mods")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, mod, "write", "r/boards:gnolang/1") // new
	shouldNotHasPerm(t, dir, usr, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/boards:gnolang/1")

	// adding "read" permission for everyone.
	dir.AddGroupPerm(Everyone, "read", ".*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, mod, "write", "r/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/boards:gnolang/1")
	shouldHasPerm(t, dir, mod, "read", "r/boards:gnolang/1") // new
	shouldHasPerm(t, dir, usr, "read", "r/boards:gnolang/1") // new
	shouldHasPerm(t, dir, cst, "read", "r/boards:gnolang/1") // new
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
		t.Errorf("%q should has perm for %q - %q", addr.String(), verb, resource)
	}
}

func shouldNotHasPerm(t *testing.T, dir *Directory, addr std.Address, verb string, resource string) {
	t.Helper()
	check := dir.HasPerm(addr, verb, resource)
	if check {
		t.Errorf("%q should not has perm for %q - %q", addr.String(), verb, resource)
	}
}
