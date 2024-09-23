package zteams

import (
    "std"
    "gno.land/p/demo/avl"
    "gno.land/p/demo/users"
)

// Team represents a team with its details and roles.
type Team struct {
    Address     std.Address
    TeamName    string
    TeamProfile string
    Owner       std.Address
    Admins      map[string]bool // Changed to string keys
    Disbursers  map[string]bool // Changed to string keys
}

// TeamManager manages teams and their members.
type TeamManager struct {
    Teams   *avl.Tree // Maps team address to Team details
    Members *avl.Tree // Maps team address to a slice of user pointers
}

// NewTeamManager initializes a new TeamManager.
func NewTeamManager() *TeamManager {
    return &TeamManager{
        Teams:   avl.NewTree(),
        Members: avl.NewTree(),
    }
}

// Helper function to retrieve a team and check for errors.
func (tm *TeamManager) getTeam(teamAddress std.Address) (*Team, error) {
    team, found := tm.Teams.Get(teamAddress.String())
    if !found {
        return nil, ErrTeamNotFound
    }
    return team.(*Team), nil
}

// Helper function to check if a user is the Owner of a team.
func (tm *TeamManager) isOwner(team *Team, userAddress std.Address) bool {
    return team.Owner == userAddress
}

// Helper function to check if a user is an Admin of a team.
func (tm *TeamManager) isAdmin(team *Team, userAddress std.Address) bool {
    return team.Admins[userAddress.String()]
}

// Designates a user as an Admin; only the Owner can do this.
func (tm *TeamManager) DesignateAdmin(teamAddress std.Address, userAddress std.Address, requester std.Address) error {
    team, err := tm.getTeam(teamAddress)
    if err != nil {
        return err
    }

    if !tm.isOwner(team, requester) {
        return ErrNotOwner
    }

    team.Admins[userAddress.String()] = true
    tm.Teams.Set(teamAddress.String(), team)
    return nil
}

// Removes an Admin; only the Owner can do this.
func (tm *TeamManager) RemoveAdmin(teamAddress std.Address, userAddress std.Address, requester std.Address) error {
    team, err := tm.getTeam(teamAddress)
    if err != nil {
        return err
    }

    if !tm.isOwner(team, requester) {
        return ErrNotOwner
    }

    delete(team.Admins, userAddress.String())
    tm.Teams.Set(teamAddress.String(), team)
    return nil
}

// Assigns disburse permission to a user; only the Owner or Admin can do this.
func (tm *TeamManager) AssignDisbursePermission(teamAddress std.Address, userAddress std.Address, requester std.Address) error {
    team, err := tm.getTeam(teamAddress)
    if err != nil {
        return err
    }

    // Only Owner or Admin can assign disburse permissions.
    if !tm.isOwner(team, requester) && !tm.isAdmin(team, requester) {
        return ErrNotAuthorized
    }

    // Assign disburse permission.
    team.Disbursers[userAddress.String()] = true
    tm.Teams.Set(teamAddress.String(), team)
    return nil
}

// Removes disburse permission from a user; only the Owner or Admin can do this.
func (tm *TeamManager) RemoveDisbursePermission(teamAddress std.Address, userAddress std.Address, requester std.Address) error {
    team, err := tm.getTeam(teamAddress)
    if err != nil {
        return err
    }

    // Only Owner or Admin can remove disburse permissions.
    if !tm.isOwner(team, requester) && !tm.isAdmin(team, requester) {
        return ErrNotAuthorized
    }

    // Check if the user has disburse permission.
    if !team.Disbursers[userAddress.String()] {
        return ErrUserNotDisburser // Use a specific error message.
    }

    // Remove disburse permission.
    delete(team.Disbursers, userAddress.String())
    tm.Teams.Set(teamAddress.String(), team)
    return nil
}

// Checks if a user can disburse (either Owner, Admin, or a designated Disburser).
func (tm *TeamManager) CanDisburse(teamAddress std.Address, userAddress std.Address) bool {
    team, err := tm.getTeam(teamAddress)
    if err != nil {
        return false
    }
    return tm.isOwner(team, userAddress) || tm.isAdmin(team, userAddress) || team.Disbursers[userAddress.String()]
}

// Adds a team to the TeamManager.
func (tm *TeamManager) AddTeam(t *Team) error {
    _, found := tm.Teams.Get(t.Address.String())
    if found {
        return ErrTeamAddressAlreadyExists
    }

    // Initialize Admins and Disbursers maps.
    t.Admins = make(map[string]bool)
    t.Disbursers = make(map[string]bool)

    tm.Teams.Set(t.Address.String(), t)
    return nil
}

// Adds a user to a team; only the Owner or Admin can do this.
func (tm *TeamManager) AddUserToTeam(teamAddress std.Address, user users.User, requester std.Address) error {
    team, err := tm.getTeam(teamAddress)
    if err != nil {
        return err
    }

    if !tm.isOwner(team, requester) && !tm.isAdmin(team, requester) {
        return ErrNotAuthorized
    }

    // Retrieve current members or initialize an empty slice if nil.
    members, _ := tm.Members.Get(teamAddress.String())
    memberList, ok := members.([]*users.User)
    if !ok || memberList == nil {
        memberList = []*users.User{}
    }

    // Check if the user is already a member.
    for _, member := range memberList {
        if member.Address == user.Address {
            return ErrUserAlreadyMember
        }
    }

    // Add user to team.
    memberList = append(memberList, &user)
    tm.Members.Set(teamAddress.String(), memberList)
    return nil
}

// Removes a user from a team; only the Owner or Admin can do this.
func (tm *TeamManager) RemoveUserFromTeam(teamAddress std.Address, userAddress std.Address, requester std.Address) error {
    team, err := tm.getTeam(teamAddress)
    if err != nil {
        return err
    }

    if !tm.isOwner(team, requester) && !tm.isAdmin(team, requester) {
        return ErrNotAuthorized
    }

    // Retrieve current members or initialize an empty slice if nil.
    members, _ := tm.Members.Get(teamAddress.String())
    memberList, ok := members.([]*users.User)
    if !ok || memberList == nil {
        return ErrUserNotMember
    }

    // Remove the user from the member list.
    userFound := false
    for i, member := range memberList {
        if member.Address == userAddress {
            memberList = append(memberList[:i], memberList[i+1:]...)
            userFound = true
            break
        }
    }

    if !userFound {
        return ErrUserNotMember
    }

    // Remove user roles and permissions.
    delete(team.Admins, userAddress.String())
    delete(team.Disbursers, userAddress.String())
    tm.Teams.Set(teamAddress.String(), team)
    tm.Members.Set(teamAddress.String(), memberList)
    return nil
}

// Gets all users in a team.
func (tm *TeamManager) GetAllUsersInTeam(teamAddress std.Address) ([]users.User, error) {
    _, err := tm.getTeam(teamAddress)
    if err != nil {
        return nil, err
    }

    // Retrieve current members or return an empty slice if nil.
    members, _ := tm.Members.Get(teamAddress.String())
    var memberList []*users.User
    if members != nil {
        memberList = members.([]*users.User)
    } else {
        return []users.User{}, nil
    }

    userList := make([]users.User, len(memberList))
    for i, member := range memberList {
        userList[i] = *member
    }

    return userList, nil
}

// Renders team information in a structured format.
func (t *Team) Render(tm *TeamManager) string {
    result := "## team " + t.TeamName + "\n\n"
    result += " * address = " + t.Address.String() + "\n"
    result += " * owner = " + t.Owner.String() + "\n\n"
    result += t.TeamProfile + "\n\n### Members\n"

    // Get the members of the team.
    members, err := tm.GetAllUsersInTeam(t.Address)
    if err != nil {
        result += "No members found.\n"
    } else {
        for _, member := range members {
            result += " * " + member.Name + " (" + member.Address.String() + "): " + member.Profile + "\n"
        }
    }

    return result
}
