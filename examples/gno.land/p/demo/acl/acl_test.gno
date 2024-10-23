package acl

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
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

func TestUserPerms(t *testing.T) {
	adm := testutils.TestAddress("admin")
	mod := testutils.TestAddress("mod")
	usr := testutils.TestAddress("user")
	cst := testutils.TestAddress("custom")

	dir := New()

	dir.AddUserPerms(usr, []string{"write", "read"}, "r/demo/boards:gnolang/2")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/2")
	shouldHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/2")
	shouldHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/2")

	dir.RemoveUserPerm(usr, "write", "r/demo/boards:gnolang/2")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/2")
	shouldHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/2")

	dir.AddUserPerm(usr, "write", "r/demo/boards:gnolang/2")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/2")
	shouldHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/2")
	shouldHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/2")

	dir.RemoveUserPerms(usr, []string{"write", "read"}, "r/demo/boards:gnolang/2")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/2")

	dir.AddUserPerms(usr, []string{"write", "read"}, "r/demo/boards:gnolang/2")
	dir.ResetUserPerms(usr)

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/2")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/2")
}

func TestUserGroups(t *testing.T) {
	adm := testutils.TestAddress("admin")
	mod := testutils.TestAddress("mod")
	usr := testutils.TestAddress("user")
	cst := testutils.TestAddress("custom")

	dir := New()

	dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	dir.AddUserToGroup(usr, "mods")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.RemoveUserFromGroup(usr, "mods")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.AddGroupPerms("admins", []string{"read", "write"}, "r/demo/boards:gnolang/3")
	dir.AddUserToGroups(adm, []string{"mods", "admins"})

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.AddGroupPerms("mods", []string{"read", "write"}, "r/demo/boards:gnolang/3")
	dir.RemoveUserFromGroups(adm, []string{"mods"})

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.RemoveGroupPerms("admins", []string{"read", "write"}, "r/demo/boards:gnolang/3")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.AddGroupPerms("admins", []string{"read", "write"}, "r/demo/boards:gnolang/3")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.RemoveGroupPerm("admins", "write", "r/demo/boards:gnolang/3")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.ResetGroupPerms("admins")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.AddGroupPerm("admins", "write", "r/demo/boards:gnolang/3")

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")

	dir.ResetUserGroups(adm)

	shouldNotHasRole(t, dir, adm, "foo")
	shouldNotHasRole(t, dir, mod, "foo")
	shouldNotHasRole(t, dir, usr, "foo")
	shouldNotHasRole(t, dir, cst, "foo")
	shouldNotHasPerm(t, dir, adm, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, adm, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, mod, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, usr, "read", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, cst, "read", "r/demo/boards:gnolang/3")
}

func TestPermDuplication(t *testing.T) {
	user := testutils.TestAddress("user")

	dir := New()
	dir.AddUserPerm(user, "write", "r/demo/boards:gnolang/3")
	dir.AddUserPerm(user, "write", "r/demo/boards:gnolang/3")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	dir.RemoveUserPerm(user, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	dir.AddUserPerms(user, []string{"write", "write"}, "r/demo/boards:gnolang/3")
	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	dir.RemoveUserPerm(user, "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	dir.AddUserToGroup(user, "mods")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	dir.RemoveGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	dir.AddGroupPerms("mods", []string{"write", "write"}, "r/demo/boards:gnolang/3")
	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	dir.RemoveGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")
}

func TestUserDuplication(t *testing.T) {
	user := testutils.TestAddress("user")

	dir := New()
	dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/4")
	dir.AddUserToGroup(user, "mods")
	dir.AddUserToGroup(user, "mods")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/4")

	dir.RemoveUserFromGroup(user, "mods")
	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/4")

	dir.AddUserToGroups(user, []string{"mods", "mods"})
	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/4")

	dir.RemoveUserFromGroup(user, "mods")
	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/4")
}

func shouldHasRole(t *testing.T, dir *Directory, addr std.Address, role string) {
	t.Helper()
	check := dir.HasRole(addr, role)
	uassert.Equal(t, true, check, ufmt.Sprintf("%s should has role %s", addr.String(), role))
}

func shouldNotHasRole(t *testing.T, dir *Directory, addr std.Address, role string) {
	t.Helper()
	check := dir.HasRole(addr, role)
	uassert.Equal(t, false, check, ufmt.Sprintf("%s should not has role %s", addr.String(), role))
}

func shouldHasPerm(t *testing.T, dir *Directory, addr std.Address, verb string, resource string) {
	t.Helper()
	check := dir.HasPerm(addr, verb, resource)
	uassert.Equal(t, true, check, ufmt.Sprintf("%s should has perm for %s - %s", addr.String(), verb, resource))
}

func shouldNotHasPerm(t *testing.T, dir *Directory, addr std.Address, verb string, resource string) {
	t.Helper()
	check := dir.HasPerm(addr, verb, resource)
	uassert.Equal(t, false, check, ufmt.Sprintf("%s should not has perm for %s - %s", addr.String(), verb, resource))
}
