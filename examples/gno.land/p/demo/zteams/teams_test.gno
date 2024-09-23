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
	}
}

// Helper function to create a sample team.
func createTeam(address, owner string) *Team {
	return &Team{
		Address: std.Address(address),
		TeamName: "Test Team",
		TeamProfile: "This is a test team profile",
		Owner: std.Address(owner),
		Admins: make(map[std.Address]bool),
		Deployers: make(map[std.Address]bool),
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
	tm.AddTeam(team)

	user := createUser("user1")

	// Only the owner or an admin can add users, so the owner must be the requester.
	err := tm.AddUserToTeam(team.Address, user, team.Owner)
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
	tm.AddTeam(team)

	user := createUser("user1")
	tm.AddUserToTeam(team.Address, user, team.Owner)

	// Remove the user.
	err := tm.RemoveUserFromTeam(team.Address, user.Address, team.Owner)
	if err != nil {
		t.Fatalf("Unexpected error when removing a user from a team: %v", err)
	}

	// Check if the user was removed.
	members, _ := tm.GetAllUsersInTeam(team.Address)
	if len(members) != 0 {
		t.Fatalf("Expected no members in the team after removal")
	}
}

/*
// Test assigning deploy permission to a user.
func TestAssignDeployPermission(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	tm.AddTeam(team)

	admin := createUser("admin1")
	user := createUser("user1")
	tm.AddUserToTeam(team.Address, admin, team.Owner)
	tm.AddUserToTeam(team.Address, user, team.Owner)

	
	// Designate the admin role to the user.
	tm.DesignateAdmin(team.Address, admin.Address)
	
	// Admin assigns deploy permission to the user.
	err := tm.AssignDeployPermission(team.Address, user.Address, admin.Address)
	if err != nil {
		t.Fatalf("Unexpected error when assigning deploy permission: %v", err)
	}
	
	// Check if the user has deploy permission.
	if !tm.CanDeploy(team.Address, user.Address) {
		t.Fatalf("Expected user to have deploy permission")
	}
}

// Test removing deploy permission from a user.
func TestRemoveDeployPermission(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	tm.AddTeam(team)

	admin := createUser("admin1")
	user := createUser("user1")
	tm.AddUserToTeam(team.Address, admin, team.Owner)
	tm.AddUserToTeam(team.Address, user, team.Owner)

	// Designate the admin role and assign deploy permission.
	tm.DesignateAdmin(team.Address, admin.Address)
	tm.AssignDeployPermission(team.Address, user.Address, admin.Address)

	// Remove deploy permission.
	err := tm.RemoveDeployPermission(team.Address, user.Address, admin.Address)
	if err != nil {
		t.Fatalf("Unexpected error when removing deploy permission: %v", err)
	}

	// Check if the user no longer has deploy permission.
	if tm.CanDeploy(team.Address, user.Address) {
		t.Fatalf("Expected user to not have deploy permission")
	}
}

// Test CanDeploy functionality for Owner, Admin, and assigned Deployer.
func TestCanDeploy(t *testing.T) {
	tm := NewTeamManager()
	team := createTeam("team1", "owner1")
	tm.AddTeam(team)

	admin := createUser("admin1")
	deployer := createUser("deployer1")
	nonDeployer := createUser("nonDeployer1")

	// Add users to the team.
	tm.AddUserToTeam(team.Address, admin, team.Owner)
	tm.AddUserToTeam(team.Address, deployer, team.Owner)
	tm.AddUserToTeam(team.Address, nonDeployer, team.Owner)

	// Assign roles and deploy permissions.
	tm.DesignateAdmin(team.Address, admin.Address)
	tm.AssignDeployPermission(team.Address, deployer.Address, admin.Address)

	// Check deploy capabilities.
	if !tm.CanDeploy(team.Address, team.Owner) {
		t.Fatalf("Expected Owner to be able to deploy")
	}

	if !tm.CanDeploy(team.Address, admin.Address) {
		t.Fatalf("Expected Admin to be able to deploy")
	}

	if !tm.CanDeploy(team.Address, deployer.Address) {
		t.Fatalf("Expected designated Deployer to be able to deploy")
	}

	if tm.CanDeploy(team.Address, nonDeployer.Address) {
		t.Fatalf("Expected non-Deployer to not be able to deploy")
	}
}*/
	
