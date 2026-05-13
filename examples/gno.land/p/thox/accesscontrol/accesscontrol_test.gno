package accesscontrol

import (
	"testing"

	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

var (
	admin    = testutils.TestAddress("admin1")
	newAdmin = testutils.TestAddress("admin2")
	user1    = testutils.TestAddress("user1")
	user2    = testutils.TestAddress("user2")

	roleName = "TestRole"
)

func initSetup(admin address) *Roles {
	return &Roles{
		Roles:   []*Role{},
		Ownable: ownable.NewWithAddress(admin),
	}
}

func TestCreateRole(t *testing.T) {
	roles := initSetup(admin)

	role, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)
	uassert.True(t, role != nil, "role should not be nil")
	uassert.Equal(t, role.Name, roleName)

	_, err = roles.CreateRole(admin, roleName)
	uassert.Error(t, err, "should fail on duplicate role creation")
}

func TestGrantRole(t *testing.T) {
	roles := initSetup(admin)

	_, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)

	err = roles.GrantRole(admin, roleName, user1)
	uassert.NoError(t, err)

	role, err := roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.True(t, role.HasAccount(user1), "user1 should have the TestRole")

	rolesList, found := roles.UserToRoles.Get(user1.String())
	uassert.True(t, found, "user1 should be in UserToRoles")
	uassert.True(t, containsRole(rolesList.([]*Role), role), "UserToRoles should contain TestRole for user1")
}

func TestGrantRoleByNonOwner(t *testing.T) {
	roles := initSetup(admin)

	_, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)

	// user2 is not the role admin (admin is); GrantRole should fail.
	err = roles.GrantRole(user2, roleName, user1)
	uassert.Error(t, err, "non-owner should not be able to grant roles")
}

func TestRevokeRole(t *testing.T) {
	roles := initSetup(admin)

	_, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)
	err = roles.GrantRole(admin, roleName, user1)
	uassert.NoError(t, err)

	err = roles.RevokeRole(admin, roleName, user1)
	uassert.NoError(t, err)

	role, err := roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.False(t, role.HasAccount(user1), "user1 should no longer have the TestRole")

	rolesList, found := roles.UserToRoles.Get(user1.String())
	if found {
		uassert.False(t, containsRole(rolesList.([]*Role), role), "UserToRoles should not contain TestRole for user1 after revocation")
	}
}

func TestRenounceRole(t *testing.T) {
	roles := initSetup(admin)

	role, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)
	err = roles.GrantRole(admin, roleName, user1)
	uassert.NoError(t, err)

	// Pas besoin de transférer la propriété pour renoncer à un rôle
	err = roles.RenounceRole(user1, roleName)
	uassert.NoError(t, err)

	role, err = roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.False(t, role.HasAccount(user1), "user1 should have renounced the TestRole")
}

func TestSetRoleAdmin(t *testing.T) {
	roles := initSetup(admin)

	role, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)

	err = role.SetRoleAdmin(admin, newAdmin)
	uassert.NoError(t, err, "admin change should succeed")

	uassert.Equal(t, role.Ownable.Owner(), newAdmin, "the new admin should be newAdmin")
	uassert.NotEqual(t, role.Ownable.Owner(), admin, "the old admin should no longer be the owner")
}

func TestCreateRoleInvalidName(t *testing.T) {
	roles := initSetup(admin)

	_, err := roles.CreateRole(admin, "")
	uassert.Error(t, err, "should fail on empty role name")

	longRoleName := "thisisaverylongrolenamethatexceedsthenormallimitfortestingpurposes"
	_, err = roles.CreateRole(admin, longRoleName)
	uassert.Error(t, err, "should fail on very long role name")
}

func TestRevokeRoleByNonOwner(t *testing.T) {
	roles := initSetup(admin)

	_, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)
	err = roles.GrantRole(admin, roleName, user1)
	uassert.NoError(t, err)

	err = roles.RevokeRole(user2, roleName, user1)
	uassert.Error(t, err, "non-owner should not be able to revoke roles")
}

func TestGrantRoleToNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	err := roles.GrantRole(admin, "NonExistentRole", user1)
	uassert.Error(t, err, "should fail when granting non-existent role")
}

func TestRevokeRoleFromNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	err := roles.RevokeRole(admin, "NonExistentRole", user1)
	uassert.Error(t, err, "should fail when revoking non-existent role")
}

func TestRenounceNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	err := roles.RenounceRole(user1, "NonExistentRole")
	uassert.Error(t, err, "should fail when renouncing non-existent role")
}

func TestDeleteRole(t *testing.T) {
	roles := initSetup(admin)

	role, err := roles.CreateRole(admin, roleName)
	uassert.NoError(t, err)
	uassert.True(t, role != nil, "role should not be nil")

	roles.Roles = []*Role{} // Clear roles for testing purpose
	_, err = roles.FindRole(roleName)
	uassert.Error(t, err, "should fail when trying to find deleted role")
}

func TestUserToRolesWithMultipleRoles(t *testing.T) {
	roles := initSetup(admin)

	roleName1 := "Role1"
	roleName2 := "Role2"

	// Create two roles
	_, err := roles.CreateRole(admin, roleName1)
	uassert.NoError(t, err)
	_, err = roles.CreateRole(admin, roleName2)
	uassert.NoError(t, err)

	// Grant both roles to user1
	err = roles.GrantRole(admin, roleName1, user1)
	uassert.NoError(t, err)
	err = roles.GrantRole(admin, roleName2, user1)
	uassert.NoError(t, err)

	// Check if user1 has both roles
	rolesList, found := roles.UserToRoles.Get(user1.String())
	uassert.True(t, found, "user1 should be in UserToRoles")
	role1, _ := roles.FindRole(roleName1)
	role2, _ := roles.FindRole(roleName2)
	uassert.True(t, containsRole(rolesList.([]*Role), role1), "UserToRoles should contain Role1 for user1")
	uassert.True(t, containsRole(rolesList.([]*Role), role2), "UserToRoles should contain Role2 for user1")
}

// func test for check if a role is in a list of roles
func containsRole(roles []*Role, target *Role) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}
