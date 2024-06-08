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
	roleData := NewRoleData("admin", admin)

	// Check initial admin role
	if roleData.AdminRole != admin {
		t.Fatalf("expected admin role to be %s, got %s", admin.String(), roleData.AdminRole.String())
	}

	// Grant role to user1
	std.TestSetOrigCaller(admin)
	roleData.GrantRole(user1)
	if !roleData.HasRole(user1) {
		t.Fatalf("expected user1 to have role")
	}

	// Check that user2 does not have the role
	if roleData.HasRole(user2) {
		t.Fatalf("expected user2 not to have role")
	}

	// Revoke role from user1
	roleData.RevokeRole(user1)
	if roleData.HasRole(user1) {
		t.Fatalf("expected user1 not to have role after revocation")
	}

	// Grant role to user1 again
	roleData.GrantRole(user1)

	// User1 renounces the role
	std.TestSetOrigCaller(user1)
	roleData.RenounceRole(user1)
	if roleData.HasRole(user1) {
		t.Fatalf("expected user1 not to have role after renouncing")
	}

	// Change admin role to user2
	std.TestSetOrigCaller(admin)
	roleData.SetRoleAdmin(user2)
	if roleData.AdminRole != user2 {
		t.Fatalf("expected admin role to be %s, got %s", user2.String(), roleData.AdminRole.String())
	}

	// User1 (now not admin) tries to grant role to user2, should panic
	std.TestSetOrigCaller(user1)
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when non-admin tries to grant role")
		}
	}()
	roleData.GrantRole(user2)
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
func TestAssertOrigCallerIsAdmin(t *testing.T) {
	adminRole := testutils.TestAddress("admin-address")
	roleData := NewRoleData("ExampleRole", adminRole)
	// Call AssertOrigCallerIsAdmin with admin caller
	std.TestSetOrigCaller(adminRole)
	err := roleData.AssertOrigCallerIsAdmin()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Call AssertOrigCallerIsAdmin with non-admin caller
	user := testutils.TestAddress("user-address")
	std.TestSetOrigCaller(user)
	err = roleData.AssertOrigCallerIsAdmin()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
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
func TestRenounceRole(t *testing.T) {
	adminRole := testutils.TestAddress("admin-address")
	roleData := NewRoleData("ExampleRole", adminRole)
	account := testutils.TestAddress("user-address")
	// Simulate the administrator as the current caller
	std.TestSetOrigCaller(adminRole)
	// Grant role to the account
	roleData.GrantRole(account)
	// Simulate the account as the current caller
	std.TestSetOrigCaller(account)
	// Renounce the role
	err := roleData.RenounceRole(account)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Check if the account still has the role
	hasRole := roleData.HasRole(account)
	if hasRole {
		t.Fatalf("expected account not to have role after renouncing")
	}
}

// Testing the SetRoleAdmin method with a New Administrator Address
func TestSetRoleAdmin(t *testing.T) {
	admin := testutils.TestAddress("admin")
	newAdmin := testutils.TestAddress("newAdmin")
	user := testutils.TestAddress("user")

	// Create role data with the administrator
	roleData := NewRoleData("admin", admin)

	// Check that the initial administrator is correct
	if roleData.AdminRole != admin {
		t.Fatalf("expected initial admin to be %s, got %s", admin.String(), roleData.AdminRole.String())
	}

	// Simulate admin as current caller
	std.TestSetOrigCaller(admin)

	// Change administrator
	roleData.SetRoleAdmin(newAdmin)

	// Check that the new administrator is correct
	if roleData.AssertOrigCallerIsAdmin() != newAdmin {
		t.Fatalf("expected new admin to be %s, got %s", newAdmin.String(), roleData.AdminRole.String())
	}

	// Simulate newAdmin as the current caller
	std.TestSetOrigCaller(newAdmin)
	roleData.AssertOrigCallerIsAdmin() // Should not panic

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
