package accesscontrol

import (
	"std"
	"testing"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

var (
	admin    = testutils.TestAddress("admin1")
	newAdmin = testutils.TestAddress("admin2")
	user1    = testutils.TestAddress("user1")
	user2    = testutils.TestAddress("user2")

	roleName = "TestRole"
)

func initSetup(admin std.Address) *Roles {
	return &Roles{
		AllRoles: []*Role{},
		Ownable:  ownable.NewWithAddress(admin),
	}
}

func TestCreateRole(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	role, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)
	uassert.True(t, role != nil, "role should not be nil")
	uassert.Equal(t, role.Name, roleName)

	_, err = roles.CreateRole(roleName)
	uassert.Error(t, err, "should fail on duplicate role creation")
}

func TestGrantRole(t *testing.T) {
	roles := initSetup(admin)

	_, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)

	err = roles.GrantRole(roleName, user1)
	uassert.NoError(t, err)

	role, err := roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.True(t, role.HasRole(user1), "user1 should have the TestRole")
}

func TestGrantRoleByNonOwner(t *testing.T) {
	roles := initSetup(admin)

	_, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)

	std.TestSetOrigCaller(user2)
	roles.Ownable.TransferOwnership(user2)
	err = roles.GrantRole(roleName, user1)
	uassert.Error(t, err, "non-owner should not be able to grant roles")

	roles.Ownable.TransferOwnership(admin)
}

func TestRevokeRole(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	_, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)
	err = roles.GrantRole(roleName, user1)
	uassert.NoError(t, err)

	err = roles.RevokeRole(roleName, user1)
	uassert.NoError(t, err)

	role, err := roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.False(t, role.HasRole(user1), "user1 should no longer have the TestRole")
}

func TestRenounceRole(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	_, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)
	err = roles.GrantRole(roleName, user1)
	uassert.NoError(t, err)

	roles.Ownable.TransferOwnership(user1)
	std.TestSetOrigCaller(user1)
	err = roles.RenounceRole(roleName)
	uassert.NoError(t, err)

	role, err := roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.False(t, role.HasRole(user1), "user1 should have renounced the TestRole")
}

func TestSetRoleAdmin(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	role, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)

	err = role.SetRoleAdmin(newAdmin)
	uassert.NoError(t, err, "admin change should succeed")

	std.TestSetOrigCaller(newAdmin)
	uassert.Equal(t, role.Ownable.Owner(), newAdmin, "the new admin should be newAdmin")

	std.TestSetOrigCaller(admin)
	uassert.NotEqual(t, role.Ownable.Owner(), admin, "the old admin should no longer be the owner")
}

func TestCreateRoleInvalidName(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	_, err := roles.CreateRole("")
	uassert.Error(t, err, "should fail on empty role name")

	longRoleName := "thisisaverylongrolenamethatexceedsthenormallimitfortestingpurposes"
	_, err = roles.CreateRole(longRoleName)
	uassert.Error(t, err, "should fail on very long role name")
}

func TestRevokeRoleByNonOwner(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	_, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)
	err = roles.GrantRole(roleName, user1)
	uassert.NoError(t, err)

	std.TestSetOrigCaller(user2)
	err = roles.RevokeRole(roleName, user1)
	uassert.Error(t, err, "non-owner should not be able to revoke roles")
}

func TestGrantRoleToNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	err := roles.GrantRole("NonExistentRole", user1)
	uassert.Error(t, err, "should fail when granting non-existent role")
}

func TestRevokeRoleFromNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	err := roles.RevokeRole("NonExistentRole", user1)
	uassert.Error(t, err, "should fail when revoking non-existent role")
}

func TestRenounceNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(user1)

	err := roles.RenounceRole("NonExistentRole")
	uassert.Error(t, err, "should fail when renouncing non-existent role")
}

func TestDeleteRole(t *testing.T) {
	roles := initSetup(admin)

	std.TestSetOrigCaller(admin)

	role, err := roles.CreateRole(roleName)
	uassert.NoError(t, err)

	roles.AllRoles = []*Role{} // Clear roles for testing purpose
	_, err = roles.FindRole(roleName)
	uassert.Error(t, err, "should fail when trying to find deleted role")
}
