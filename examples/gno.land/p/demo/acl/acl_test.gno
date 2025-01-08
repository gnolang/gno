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
	res := dir.AddUserPerm(adm, ".*", ".*")
	uassert.True(t, res, "failed to add perm")

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
	res2 := dir.AddUserPerm(cst, "write", "r/demo/boards:gnolang/.*")
	uassert.True(t, res2, "failed to add perm")

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
	res3 := dir.AddGroupPerm("mods", "role", "moderator")
	res4 := dir.AddGroupPerm("mods", "write", ".*")
	uassert.True(t, res3, "failed to add group perm")
	uassert.True(t, res4, "failed to add group perm")

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
	res5 := dir.AddUserToGroup(mod, "mods")
	uassert.True(t, res5, "failed to add user to group")

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
	res6 := dir.AddGroupPerm(Everyone, "read", ".*")
	uassert.True(t, res6, "failed to add group perm")

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

	res := dir.AddUserPerms(usr, []string{"write", "read"}, "r/demo/boards:gnolang/2")
	uassert.True(t, res, "failed to add user perms")

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

	res1 := dir.RemoveUserPerm(usr, "write", "r/demo/boards:gnolang/2")
	res2 := dir.RemoveUserPerm(usr, "write", "r/demo/boards:gnolang/2")
	uassert.True(t, res1, "failed to remove user perm")
	uassert.False(t, res2, "should not remove user perm")

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

	res3 := dir.AddUserPerm(usr, "write", "r/demo/boards:gnolang/2")
	uassert.True(t, res3, "failed to add user perm")

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

	res4 := dir.RemoveUserPerms(usr, []string{"write", "read"}, "r/demo/boards:gnolang/2")
	res5 := dir.RemoveUserPerms(usr, []string{"write", "read"}, "r/demo/boards:gnolang/2")
	uassert.True(t, res4, "failed to remove user perms")
	uassert.False(t, res5, "should not remove user perms")

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

	res6 := dir.AddUserPerms(usr, []string{"write", "read"}, "r/demo/boards:gnolang/2")
	res7 := dir.ResetUserPerms(usr)
	uassert.True(t, res6, "failed to add user perms")
	uassert.True(t, res7, "failed to reset user perms")

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

	res := dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	res2 := dir.AddUserToGroup(usr, "mods")
	uassert.True(t, res, "failed to add group perm")
	uassert.True(t, res2, "failed to add user to group")

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

	res3 := dir.RemoveUserFromGroup(usr, "mods")
	uassert.True(t, res3, "failed to remove user from group")

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

	res4 := dir.AddGroupPerms("admins", []string{"read", "write"}, "r/demo/boards:gnolang/3")
	res5 := dir.AddUserToGroups(adm, []string{"mods", "admins"})
	uassert.True(t, res4, "failed to add group perms")
	uassert.True(t, res5, "failed to add user to groups")

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

	res6 := dir.AddGroupPerms("mods", []string{"read", "write"}, "r/demo/boards:gnolang/3")
	res7 := dir.RemoveUserFromGroups(adm, []string{"mods"})
	res8 := dir.RemoveUserFromGroups(adm, []string{"mods"})
	uassert.True(t, res6, "failed to add group perms")
	uassert.True(t, res7, "failed to remove user from groups")
	uassert.False(t, res8, "should not remove user from groups")

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

	res9 := dir.RemoveGroupPerms("admins", []string{"read", "write"}, "r/demo/boards:gnolang/3")
	res10 := dir.RemoveGroupPerms("admins", []string{"read", "write"}, "r/demo/boards:gnolang/3")
	uassert.True(t, res9, "failed to remove group perms")
	uassert.False(t, res10, "should not remove group perms")

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

	res11 := dir.AddGroupPerms("admins", []string{"read", "write"}, "r/demo/boards:gnolang/3")
	uassert.True(t, res11, "failed to add group perms")

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

	res12 := dir.RemoveGroupPerm("admins", "write", "r/demo/boards:gnolang/3")
	res13 := dir.RemoveGroupPerm("admins", "write", "r/demo/boards:gnolang/3")
	uassert.True(t, res12, "failed to remove group perm")
	uassert.False(t, res13, "should not remove group perm")

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

	res14 := dir.ResetGroupPerms("admins")
	uassert.True(t, res14, "failed to reset group perms")

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

	res15 := dir.AddGroupPerm("admins", "write", "r/demo/boards:gnolang/3")
	uassert.True(t, res15, "failed to add group perm")

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

	res16 := dir.ResetUserGroups(adm)
	uassert.True(t, res16, "failed to reset user groups")

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
	res := dir.AddUserPerm(user, "write", "r/demo/boards:gnolang/3")
	res1 := dir.AddUserPerm(user, "write", "r/demo/boards:gnolang/3")
	uassert.True(t, res, "failed to add user perm")
	uassert.True(t, res1, "failed to add user perm")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	res2 := dir.RemoveUserPerm(user, "write", "r/demo/boards:gnolang/3")
	uassert.True(t, res2, "failed to add user perm")

	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	res3 := dir.AddUserPerms(user, []string{"write", "write"}, "r/demo/boards:gnolang/3")
	uassert.True(t, res3, "failed to add user perm")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	res4 := dir.RemoveUserPerm(user, "write", "r/demo/boards:gnolang/3")
	uassert.True(t, res4, "failed to add user perm")

	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	res5 := dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	res6 := dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	res7 := dir.AddUserToGroup(user, "mods")
	uassert.True(t, res5, "failed to add group perm")
	uassert.True(t, res6, "failed to add group perm")
	uassert.True(t, res7, "failed to add user to group")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	res8 := dir.RemoveGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	uassert.True(t, res8, "failed to add group perm")

	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	res9 := dir.AddGroupPerms("mods", []string{"write", "write"}, "r/demo/boards:gnolang/3")
	uassert.True(t, res9, "failed to add group perm")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")

	res10 := dir.RemoveGroupPerm("mods", "write", "r/demo/boards:gnolang/3")
	uassert.True(t, res10, "failed to add group perm")

	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/3")
}

func TestUserDuplication(t *testing.T) {
	user := testutils.TestAddress("user")

	dir := New()
	res := dir.AddGroupPerm("mods", "write", "r/demo/boards:gnolang/4")
	res1 := dir.AddUserToGroup(user, "mods")
	res2 := dir.AddUserToGroup(user, "mods")
	uassert.True(t, res, "failed to add group perm")
	uassert.True(t, res1, "failed to add user to group")
	uassert.True(t, res2, "failed to add user to group")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/4")

	res3 := dir.RemoveUserFromGroup(user, "mods")
	uassert.True(t, res3, "failed to remove user from group")

	shouldNotHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/4")

	res4 := dir.AddUserToGroups(user, []string{"mods", "mods"})
	uassert.True(t, res4, "failed to add user to groups")

	shouldHasPerm(t, dir, user, "write", "r/demo/boards:gnolang/4")

	res5 := dir.RemoveUserFromGroup(user, "mods")
	res6 := dir.RemoveUserFromGroup(user, "mods")
	uassert.True(t, res5, "failed to remove user from group")
	uassert.False(t, res6, "should not remove user from group")

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
