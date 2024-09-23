package zteams

import (
	"std"
	"gno.land/p/demo/users"
	"testing"
)

// Helper function to create a sample user.
func createUser(address string) users.User {
	return users.User{
		Address: std.Address(address),
		Name:    "Test User",
		Profile: "Sample profile",
		Number:  1,
		Invites: 1,
		Inviter: std.Address("inviter"),
	}
}

// Helper function to create a sample team.
func createTeam(address, owner string) *Team {
	return &Team{
		Address:     std.Address(address),
		TeamName:    "Test Team",
		TeamProfile: "This is a test team profile",
		Owner:       std.Address(owner),
		Admins:      make(map[string]bool), // Updated to use string keys
		Disbursers:  make(map[string]bool), // Updated to use string keys
	}
}

// Test creating a new TeamManager.
func TestNewTeamManager(t *testing.T) {
	tm := NewTeamManager()
	if tm.Teams == nil || tm.Members == nil {
		t.Fatalf("Expected Teams and Members to be initialized")
	}
}

// Test adding a team.
func TestAddTeam(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")

	err := tm.AddTeam(team)
	if err != nil {
		t.Fatalf("Unexpected error when adding a team: %v", err)
	}

	// Try adding the same team again to check for duplicate error.
	err = tm.AddTeam(team)
	if err == nil {
		t.Fatalf("Expected an error when adding a duplicate team")
	}
}

// Test adding a user to a team.
func TestAddUserToTeam(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	err := tm.AddTeam(team)
	if err != nil {
		t.Fatalf("Unexpected error when adding a team: %v", err)
	}

	user := createUser("user1")

	// Only the owner or an admin can add users, so the owner must be the requester.
	err = tm.AddUserToTeam(team.Address, user, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding a user to a team: %v", err)
	}

	// Check if the user was added.
	members, _ := tm.GetAllUsersInTeam(team.Address)
	if len(members) != 1 || members[0].Address != user.Address {
		t.Fatalf("Expected user to be added to the team")
	}
}

// Test removing a user from a team.
func TestRemoveUserFromTeam(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	err := tm.AddTeam(team)
	if err != nil {
		t.Fatalf("Unexpected error when adding a team: %v", err)
	}

	user := createUser("user1")
	err = tm.AddUserToTeam(team.Address, user, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding a user to a team: %v", err)
	}

	// Remove the user.
	err = tm.RemoveUserFromTeam(team.Address, user.Address, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when removing a user from a team: %v", err)
	}

	// Check if the user was removed.
	members, _ := tm.GetAllUsersInTeam(team.Address)
	if len(members) != 0 {
		t.Fatalf("Expected no members in the team after removal")
	}
}

// Test assigning disburse permission to a user.
func TestAssignDisbursePermission(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	err := tm.AddTeam(team)
	if err != nil {
		t.Fatalf("Unexpected error when adding a team: %v", err)
	}

	admin := createUser("admin1")
	user := createUser("user1")
	err = tm.AddUserToTeam(team.Address, admin, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding admin to the team: %v", err)
	}
	err = tm.AddUserToTeam(team.Address, user, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding user to the team: %v", err)
	}

	// Designate the admin role to the admin user.
	err = tm.DesignateAdmin(team.Address, admin.Address, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when designating admin: %v", err)
	}

	// Admin assigns disburse permission to the user.
	err = tm.AssignDisbursePermission(team.Address, user.Address, admin.Address)
	if err != nil {
		t.Fatalf("Unexpected error when assigning disburse permission: %v", err)
	}

	// Check if the user has disburse permission.
	if !tm.CanDisburse(team.Address, user.Address) {
		t.Fatalf("Expected user to have disburse permission")
	}
}

// Test removing disburse permission from a user.
func TestRemoveDisbursePermission(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	err := tm.AddTeam(team)
	if err != nil {
		t.Fatalf("Unexpected error when adding a team: %v", err)
	}

	admin := createUser("admin1")
	user := createUser("user1")
	err = tm.AddUserToTeam(team.Address, admin, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding admin to the team: %v", err)
	}
	err = tm.AddUserToTeam(team.Address, user, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding user to the team: %v", err)
	}

	// Designate the admin role and assign disburse permission.
	err = tm.DesignateAdmin(team.Address, admin.Address, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when designating admin: %v", err)
	}
	err = tm.AssignDisbursePermission(team.Address, user.Address, admin.Address)
	if err != nil {
		t.Fatalf("Unexpected error when assigning disburse permission: %v", err)
	}

	// Remove disburse permission.
	err = tm.RemoveDisbursePermission(team.Address, user.Address, admin.Address)
	if err != nil {
		t.Fatalf("Unexpected error when removing disburse permission: %v", err)
	}

	// Check if the user no longer has disburse permission.
	if tm.CanDisburse(team.Address, user.Address) {
		t.Fatalf("Expected user to not have disburse permission")
	}
}

// Test CanDisburse functionality for Owner, Admin, and assigned Disburser.
func TestCanDisburse(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	err := tm.AddTeam(team)
	if err != nil {
		t.Fatalf("Unexpected error when adding a team: %v", err)
	}

	admin := createUser("admin1")
	disburser := createUser("disburser1")
	nonDisburser := createUser("nonDisburser1")

	// Add users to the team.
	err = tm.AddUserToTeam(team.Address, admin, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding admin to the team: %v", err)
	}
	err = tm.AddUserToTeam(team.Address, disburser, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding disburser to the team: %v", err)
	}
	err = tm.AddUserToTeam(team.Address, nonDisburser, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when adding nonDisburser to the team: %v", err)
	}

	// Assign roles and disburse permissions.
	err = tm.DesignateAdmin(team.Address, admin.Address, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when designating admin: %v", err)
	}
	err = tm.AssignDisbursePermission(team.Address, disburser.Address, admin.Address)
	if err != nil {
		t.Fatalf("Unexpected error when assigning disburse permission: %v", err)
	}

	// Check disburse capabilities.
	if !tm.CanDisburse(team.Address, team.Owner) {
		t.Fatalf("Expected Owner to be able to disburse")
	}

	if !tm.CanDisburse(team.Address, admin.Address) {
		t.Fatalf("Expected Admin to be able to disburse")
	}

	if !tm.CanDisburse(team.Address, disburser.Address) {
		t.Fatalf("Expected designated Disburser to be able to disburse")
	}

	if tm.CanDisburse(team.Address, nonDisburser.Address) {
		t.Fatalf("Expected non-Disburser to not be able to disburse")
	}
}
