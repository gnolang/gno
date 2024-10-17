package teams_test

import (
	"testing"
	"std"
	"gno.land/p/demo/users"
	"gno.land/p/demo/teams"
)

func TestCreateTeam(t *testing.T) {
	teamAddress := std.Address("g1teamaddress")
	teamName := "Test Team"
	owner := std.Address("g1owneraddress")

	err := teams.CreateTeam(teamAddress, teamName, owner)
	if err != nil {
		t.Errorf("Failed to create team: %v", err)
	}

	// Attempt to create a team with the same address, should fail
	err = teams.CreateTeam(teamAddress, teamName, owner)
	if err == nil {
		t.Errorf("Expected error when creating a duplicate team, got nil")
	}
}

func TestAddMember(t *testing.T) {
	teamAddress := std.Address("g1teamaddress")
	owner := std.Address("g1owneraddress")
	_ = teams.CreateTeam(teamAddress, "Test Team", owner)

	user := users.User{Name: "Test User", Address: std.Address("g1useraddress")}
	err := teams.AddMember(teamAddress, user)
	if err != nil {
		t.Errorf("Failed to add member: %v", err)
	}

	// Attempt to add the same member again, should fail
	err = teams.AddMember(teamAddress, user)
	if err == nil {
		t.Errorf("Expected error when adding a duplicate member, got nil")
	}
}

func TestRemoveMember(t *testing.T) {
	teamAddress := std.Address("g1teamaddress")
	owner := std.Address("g1owneraddress")
	_ = teams.CreateTeam(teamAddress, "Test Team", owner)

	user := users.User{Name: "Test User", Address: std.Address("g1useraddress")}
	_ = teams.AddMember(teamAddress, user)

	err := teams.RemoveMember(teamAddress, user.Address)
	if err != nil {
		t.Errorf("Failed to remove member: %v", err)
	}

	// Attempt to remove a non-existing member, should fail
	err = teams.RemoveMember(teamAddress, user.Address)
	if err == nil {
		t.Errorf("Expected error when removing a non-existing member, got nil")
	}
}

func TestAddPermission(t *testing.T) {
	teamAddress := std.Address("g1teamaddress")
	owner := std.Address("g1owneraddress")
	_ = teams.CreateTeam(teamAddress, "Test Team", owner)

	permission := "manage_tasks"
	err := teams.AddPermission(teamAddress, permission)
	if err != nil {
		t.Errorf("Failed to add permission: %v", err)
	}

	// Attempt to add the same permission again, should fail
	err = teams.AddPermission(teamAddress, permission)
	if err == nil {
		t.Errorf("Expected error when adding a duplicate permission, got nil")
	}
}

func TestRemovePermission(t *testing.T) {
	teamAddress := std.Address("g1teamaddress")
	owner := std.Address("g1owneraddress")
	_ = teams.CreateTeam(teamAddress, "Test Team", owner)

	permission := "manage_tasks"
	_ = teams.AddPermission(teamAddress, permission)

	err := teams.RemovePermission(teamAddress, permission)
	if err != nil {
		t.Errorf("Failed to remove permission: %v", err)
	}

	// Attempt to remove a non-existing permission, should not fail
	err = teams.RemovePermission(teamAddress, permission)
	if err != nil {
		t.Errorf("Expected no error when removing a non-existing permission, got %v", err)
	}
}

func TestAddMemberPermission(t *testing.T) {
	teamAddress := std.Address("g1teamaddress")
	owner := std.Address("g1owneraddress")
	_ = teams.CreateTeam(teamAddress, "Test Team", owner)

	user := users.User{Name: "Test User", Address: std.Address("g1useraddress")}
	_ = teams.AddMember(teamAddress, user)

	permission := "manage_tasks"
	_ = teams.AddPermission(teamAddress, permission)

	err := teams.AddMemberPermission(teamAddress, user.Address, permission)
	if err != nil {
		t.Errorf("Failed to add permission to member: %v", err)
	}

	// Attempt to add a permission that does not exist in the team, should fail
	err = teams.AddMemberPermission(teamAddress, user.Address, "non_existing_permission")
	if err == nil {
		t.Errorf("Expected error when adding a non-existing permission to member, got nil")
	}
}

func TestRemoveMemberPermission(t *testing.T) {
	teamAddress := std.Address("g1teamaddress")
	owner := std.Address("g1owneraddress")
	_ = teams.CreateTeam(teamAddress, "Test Team", owner)

	user := users.User{Name: "Test User", Address: std.Address("g1useraddress")}
	_ = teams.AddMember(teamAddress, user)

	permission := "manage_tasks"
	_ = teams.AddPermission(teamAddress, permission)
	_ = teams.AddMemberPermission(teamAddress, user.Address, permission)

	err := teams.RemoveMemberPermission(teamAddress, user.Address, permission)
	if err != nil {
		t.Errorf("Failed to remove permission from member: %v", err)
	}

	// Attempt to remove a permission that the member does not have, should not fail
	err = teams.RemoveMemberPermission(teamAddress, user.Address, permission)
	if err != nil {
		t.Errorf("Expected no error when removing a non-existing permission from member, got %v", err)
	}
}
