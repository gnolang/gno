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

func initSetup(adminAddr address) *Roles {
	return &Roles{
		Roles:   []*Role{},
		Ownable: ownable.NewWithAddress(adminAddr),
	}
}

// asAdmin / asUser1 / asUser2 wrap the migrated (_ int, rlm realm)
// API in a crossing closure under SetRealm(NewUserRealm(addr)) so
// rlm.Previous().Address() inside the callee resolves to addr.

func TestCreateRole(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var (
		role *Role
		err  error
	)
	func(cur realm) { role, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)
	uassert.True(t, role != nil, "role should not be nil")
	uassert.Equal(t, role.Name, roleName)

	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.Error(t, err, "should fail on duplicate role creation")
}

func TestGrantRole(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)

	func(cur realm) { err = roles.GrantRole(0, cur, roleName, user1) }(cross)
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

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)

	// user2 is not the role admin (admin is); GrantRole should fail.
	testing.SetRealm(testing.NewUserRealm(user2))
	func(cur realm) { err = roles.GrantRole(0, cur, roleName, user1) }(cross)
	uassert.Error(t, err, "non-owner should not be able to grant roles")
}

func TestRevokeRole(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)
	func(cur realm) { err = roles.GrantRole(0, cur, roleName, user1) }(cross)
	uassert.NoError(t, err)

	func(cur realm) { err = roles.RevokeRole(0, cur, roleName, user1) }(cross)
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

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)
	func(cur realm) { err = roles.GrantRole(0, cur, roleName, user1) }(cross)
	uassert.NoError(t, err)

	// user1 renounces their own role — must be in user1's context.
	testing.SetRealm(testing.NewUserRealm(user1))
	func(cur realm) { err = roles.RenounceRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)

	role, err := roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.False(t, role.HasAccount(user1), "user1 should have renounced the TestRole")
}

func TestRenounceRoleByNonHolder(t *testing.T) {
	// Regression for the address-parameter forgery: previously
	// RenounceRole(victim) could be called by any realm to force a
	// victim to renounce. After the IsCurrent + rlm.Previous() fix,
	// the renouncer is the principal — an attacker cannot strip
	// someone else's role.
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)
	func(cur realm) { err = roles.GrantRole(0, cur, roleName, user1) }(cross)
	uassert.NoError(t, err)

	// user2 tries to "renounce" user1's role — must fail because user2
	// does not hold the role.
	testing.SetRealm(testing.NewUserRealm(user2))
	func(cur realm) { err = roles.RenounceRole(0, cur, roleName) }(cross)
	uassert.Error(t, err, "non-holder must not be able to renounce")

	role, err := roles.FindRole(roleName)
	uassert.NoError(t, err)
	uassert.True(t, role.HasAccount(user1), "user1 must still hold the role")
}

func TestSetRoleAdmin(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var (
		role *Role
		err  error
	)
	func(cur realm) { role, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)

	func(cur realm) { err = role.SetRoleAdmin(0, cur, newAdmin) }(cross)
	uassert.NoError(t, err, "admin change should succeed")

	uassert.Equal(t, role.Ownable.Owner(), newAdmin, "the new admin should be newAdmin")
	uassert.NotEqual(t, role.Ownable.Owner(), admin, "the old admin should no longer be the owner")
}

func TestCreateRoleInvalidName(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, "") }(cross)
	uassert.Error(t, err, "should fail on empty role name")

	longRoleName := "thisisaverylongrolenamethatexceedsthenormallimitfortestingpurposes"
	func(cur realm) { _, err = roles.CreateRole(0, cur, longRoleName) }(cross)
	uassert.Error(t, err, "should fail on very long role name")
}

func TestRevokeRoleByNonOwner(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName) }(cross)
	uassert.NoError(t, err)
	func(cur realm) { err = roles.GrantRole(0, cur, roleName, user1) }(cross)
	uassert.NoError(t, err)

	testing.SetRealm(testing.NewUserRealm(user2))
	func(cur realm) { err = roles.RevokeRole(0, cur, roleName, user1) }(cross)
	uassert.Error(t, err, "non-owner should not be able to revoke roles")
}

func TestGrantRoleToNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { err = roles.GrantRole(0, cur, "NonExistentRole", user1) }(cross)
	uassert.Error(t, err, "should fail when granting non-existent role")
}

func TestRevokeRoleFromNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { err = roles.RevokeRole(0, cur, "NonExistentRole", user1) }(cross)
	uassert.Error(t, err, "should fail when revoking non-existent role")
}

func TestRenounceNonExistentRole(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(user1))
	var err error
	func(cur realm) { err = roles.RenounceRole(0, cur, "NonExistentRole") }(cross)
	uassert.Error(t, err, "should fail when renouncing non-existent role")
}

func TestDeleteRole(t *testing.T) {
	roles := initSetup(admin)

	testing.SetRealm(testing.NewUserRealm(admin))
	var (
		role *Role
		err  error
	)
	func(cur realm) { role, err = roles.CreateRole(0, cur, roleName) }(cross)
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
	testing.SetRealm(testing.NewUserRealm(admin))
	var err error
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName1) }(cross)
	uassert.NoError(t, err)
	func(cur realm) { _, err = roles.CreateRole(0, cur, roleName2) }(cross)
	uassert.NoError(t, err)

	// Grant both roles to user1
	func(cur realm) { err = roles.GrantRole(0, cur, roleName1, user1) }(cross)
	uassert.NoError(t, err)
	func(cur realm) { err = roles.GrantRole(0, cur, roleName2, user1) }(cross)
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
