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
	roleData := NewRoleData(admin)

	// Check initial admin role
	if roleData.GetRoleAdmin() != admin {
		t.Fatalf("expected admin role to be %s, got %s", admin.String(), roleData.adminRole.String())
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
	if roleData.adminRole != user2 {
		t.Fatalf("expected admin role to be %s, got %s", user2.String(), roleData.adminRole.String())
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
