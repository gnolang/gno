package acl

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/uassert"
	"gno.land/p/nt/ufmt"
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
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/1")

	// adding all the rights to admin.
	dir.AddUserPerm(adm, ".*", ".*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/1") // new
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/1") // new
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/1")

	// adding custom regexp rule for user "cst".
	dir.AddUserPerm(cst, "write", "r/demo/boards:gnolang/.*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/1") // new
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/1")

	// adding a group perm for a new group.
	// no changes expected.
	dir.AddGroupPerm("mods", "role", "moderator")
	dir.AddGroupPerm("mods", "write", ".*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/1")

	// assigning the user "mod" to the "mods" group.
	dir.AddUserToGroup(mod, "mods")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/1") // new
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/1")

	// adding "read" permission for everyone.
	dir.AddGroupPerm(Everyone, "read", ".*")
	shouldHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/1")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/1")
	shouldHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/1") // new
	shouldHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/1") // new
	shouldHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/1") // new
}

func shouldHasRole(t *testing.T, dir *Directory, addr address, role string) {
	t.Helper()
	check := dir.HasRole(addr, role)
	uassert.Equal(t, true, check, ufmt.Sprintf("%s should has role %s", addr.String(), role))
}

func shouldNotHasRole(t *testing.T, dir *Directory, addr address, role string) {
	t.Helper()
	check := dir.HasRole(addr, role)
	uassert.Equal(t, false, check, ufmt.Sprintf("%s should not has role %s", addr.String(), role))
}

func shouldHasPerm(t *testing.T, dir *Directory, addr address, verb string, resource string) {
	t.Helper()
	check := dir.HasPerm(addr, verb, resource)
	uassert.Equal(t, true, check, ufmt.Sprintf("%s should has perm for %s - %s", addr.String(), verb, resource))
}

func shouldNotHasPerm(t *testing.T, dir *Directory, addr address, verb string, resource string) {
	t.Helper()
	check := dir.HasPerm(addr, verb, resource)
	uassert.Equal(t, false, check, ufmt.Sprintf("%s should not has perm for %s - %s", addr.String(), verb, resource))
}
