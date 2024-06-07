package accesscontrol

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

// TestAccessControl verifies the access control functionality.
func TestAccessControl(t *testing.T) {
	admin := testutils.TestAddress("admin")
	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")

	// Create new RoleData
	RoleData := NewRoleData("admin", admin)

	// Check initial admin role
	if RoleData.GetRoleAdmin() != admin {
		t.Fatalf("expected admin role to be %s, got %s", admin.String(), RoleData.AdminRole.String())
	}

	// Grant role to user1
	std.TestSetOrigCaller(admin)
	RoleData.GrantRole(user1)
	if !RoleData.HasRole(user1) {
		t.Fatalf("expected user1 to have role")
	}

	// Check that user2 does not have the role
	if RoleData.HasRole(user2) {
		t.Fatalf("expected user2 not to have role")
	}

	// Revoke role from user1
	RoleData.RevokeRole(user1)
	if RoleData.HasRole(user1) {
		t.Fatalf("expected user1 not to have role after revocation")
	}

	// Grant role to user1 again
	RoleData.GrantRole(user1)

	// User1 renounces the role
	std.TestSetOrigCaller(user1)
	RoleData.RenounceRole(user1)
	if RoleData.HasRole(user1) {
		t.Fatalf("expected user1 not to have role after renouncing")
	}

	// Change admin role to user2
	std.TestSetOrigCaller(admin)
	RoleData.SetRoleAdmin(user2)
	if RoleData.AdminRole != user2 {
		t.Fatalf("expected admin role to be %s, got %s", user2.String(), RoleData.AdminRole.String())
	}

	// User1 (now not admin) tries to grant role to user2, should panic
	std.TestSetOrigCaller(user1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when non-admin tries to grant role")
		}
	}()
	RoleData.GrantRole(user2)
}

// TestCreateRole tests the CreateRole method of the RoleData struct
func TestCreateRole(t *testing.T) {
	// Simulate the administrator as the current caller
	admin := testutils.TestAddress("admin")
	std.TestSetOrigCaller(admin)

	// Create role data with the administrator
	roleData := NewRoleData("admin", admin)

	// Create a new role with a new administrator address
	newAdmin := testutils.TestAddress("newAdmin")
	newRoleName := "newRole"
	newRole := roleData.CreateRole(newRoleName, newAdmin)

	// Check that the new role has been created correctly
	if newRole.Name != newRoleName {
		t.Fatalf("expected new role name to be '%s', got '%s'", newRoleName, newRole.Name)
	}
	if newRole.AdminRole != newAdmin {
		t.Fatalf("expected new role admin role to be %s, got %s", newAdmin.String(), newRole.AdminRole.String())
	}

	// Simulate newAdmin as the current caller
	std.TestSetOrigCaller(newAdmin)

	// Explicitly add the role to the new administrator to check functionality
	newRole.GrantRole(newAdmin)

	// Check if the new role has been added to the holder
	if !newRole.HasRole(newAdmin) {
		t.Fatalf("expected new role to be added to the holder")
	}
}

// TestOnlyAdmin verifies the OnlyAdmin method.
func TestOnlyAdmin(t *testing.T) {
	admin := testutils.TestAddress("admin")
	user := testutils.TestAddress("user")
	roleData := NewRoleData("admin", admin)
	// Call OnlyAdmin method with admin caller
	std.TestSetOrigCaller(admin)
	roleData.OnlyAdmin() // Should not panic
	// Call OnlyAdmin method with non-admin caller
	std.TestSetOrigCaller(user)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when non-admin calls OnlyAdmin")
		}
	}()
	roleData.OnlyAdmin()
}

// Testing the RevokeRole Method for a Non-Admin
func TestRevokeRoleNonAdmin(t *testing.T) {
	admin := testutils.TestAddress("admin")
	user1 := testutils.TestAddress("user1")
	user2 := testutils.TestAddress("user2")

	// Create role data with the administrator
	roleData := NewRoleData("admin", admin)

	// Grant role to user1
	std.TestSetOrigCaller(admin)
	roleData.GrantRole(user1)
	if !roleData.HasRole(user1) {
		t.Fatalf("expected user1 to have role")
	}

	// Simulate user2 as the current caller
	std.TestSetOrigCaller(user2)

	// Attempting to revoke user1's role as user2 (non-admin), should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when non-admin tries to revoke role")
		}
	}()
	roleData.RevokeRole(user1)
}

// Testing the RenounceRole method with Invalid Confirmation
func TestRenounceRoleInvalidConfirmation(t *testing.T) {
	admin := testutils.TestAddress("admin")
	user := testutils.TestAddress("user")

	// Create role data with the administrator
	roleData := NewRoleData("admin", admin)

	// Grant role to user
	std.TestSetOrigCaller(admin)
	roleData.GrantRole(user)
	if !roleData.HasRole(user) {
		t.Fatalf("expected user to have role")
	}

	// Simulate user as current caller
	std.TestSetOrigCaller(user)

	// Attempting to relinquish the role with an invalid confirmation, should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when caller confirmation does not match")
		}
	}()
	roleData.RenounceRole(admin) // Pass an invalid confirmation
}

// Testing the SetRoleAdmin method with a New Administrator Address
func TestSetRoleAdmin(t *testing.T) {
	admin := testutils.TestAddress("admin")
	newAdmin := testutils.TestAddress("newAdmin")
	user := testutils.TestAddress("user")

	// Create role data with the administrator
	roleData := NewRoleData("admin", admin)

	// Check that the initial administrator is correct
	if roleData.GetRoleAdmin() != admin {
		t.Fatalf("expected initial admin to be %s, got %s", admin.String(), roleData.GetRoleAdmin().String())
	}

	// Simulate admin as current caller
	std.TestSetOrigCaller(admin)

	// Change administrator
	roleData.SetRoleAdmin(newAdmin)

	// Check that the new administrator is correct
	if roleData.GetRoleAdmin() != newAdmin {
		t.Fatalf("expected new admin to be %s, got %s", newAdmin.String(), roleData.GetRoleAdmin().String())
	}

	// Simulate newAdmin as the current caller
	std.TestSetOrigCaller(newAdmin)
	roleData.OnlyAdmin() // Should not panic

	// Add a role to a user
	roleData.GrantRole(user)
	if !roleData.HasRole(user) {
		t.Fatalf("expected user to have role")
	}

	// Simulate initial admin as current caller
	std.TestSetOrigCaller(admin)

	// Attempting to revoke a user's role by the former administrator should cause panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when former admin tries to revoke role")
		}
	}()
	roleData.RevokeRole(user)
}