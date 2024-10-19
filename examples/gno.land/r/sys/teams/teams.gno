package teams

import (
	"errors"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/users"
)

var Teams *avl.Tree
var TeamNames *avl.Tree

func init() {
	Teams = avl.NewTree()
	TeamNames = avl.NewTree()
}

func CreateTeam(teamAddress std.Address, teamName string, owner std.Address) error {
	if _, exists := Teams.Get(teamAddress.String()); exists {
		return errors.New("team already exists")
	}

	if _, nameExists := TeamNames.Get(strings.ToLower(teamName)); nameExists {
		return errors.New("team name already exists")
	}

	team := &Team{
		Name:   teamName,
		Owners: []std.Address{owner},
	}

	Teams.Set(teamAddress.String(), team)
	TeamNames.Set(strings.ToLower(teamName), teamAddress)
	return nil
}

func AddMember(teamAddress std.Address, user users.User) error {
	team, err := getTeam(teamAddress)
	if err != nil {
		return err
	}

	// Check if the user is already a member
	for _, member := range team.Members {
		if member.User.Address == user.Address {
			return errors.New("user is already a member of the team")
		}
	}

	team.Members = append(team.Members, TeamMember{User: user})
	Teams.Set(teamAddress.String(), team)
	return nil
}

func RemoveMember(teamAddress std.Address, userAddress std.Address) error {
	team, err := getTeam(teamAddress)
	if err != nil {
		return err
	}

	for i, member := range team.Members {
		if member.User.Address == userAddress {
			team.Members = append(team.Members[:i], team.Members[i+1:]...)
			Teams.Set(teamAddress.String(), team)
			return nil
		}
	}

	return errors.New("member not found")
}

func AddPermission(teamAddress std.Address, permission string) error {
	team, err := getTeam(teamAddress)
	if err != nil {
		return err
	}

	for _, perm := range team.Permissions {
		if strings.EqualFold(perm, permission) {
			return errors.New("permission already exists in the team")
		}
	}

	team.Permissions = append(team.Permissions, permission)
	Teams.Set(teamAddress.String(), team)
	return nil
}

func RemovePermission(teamAddress std.Address, permission string) error {
	team, err := getTeam(teamAddress)
	if err != nil {
		return err
	}

	for i, perm := range team.Permissions {
		if strings.EqualFold(perm, permission) {
			team.Permissions = append(team.Permissions[:i], team.Permissions[i+1:]...)
			break
		}
	}

	for i, member := range team.Members {
		for j, memberPerm := range member.Permissions {
			if strings.EqualFold(memberPerm, permission) {
				team.Members[i].Permissions = append(member.Permissions[:j], member.Permissions[j+1:]...)
				break
			}
		}
	}

	Teams.Set(teamAddress.String(), team)
	return nil
}

// AddMemberPermission adds a permission to a specific member of the team
func AddMemberPermission(teamAddress std.Address, userAddress std.Address, permission string) error {
	team, err := getTeam(teamAddress)
	if err != nil {
		return err
	}

	// Check if the permission exists in the team's permissions
	permissionExists := false
	for _, perm := range team.Permissions {
		if strings.EqualFold(perm, permission) {
			permissionExists = true
			break
		}
	}
	if !permissionExists {
		return errors.New("permission does not exist in the team")
	}

	// Find the member and add the permission if they don't already have it
	for i, member := range team.Members {
		if member.User.Address == userAddress {
			for _, memberPerm := range member.Permissions {
				if strings.EqualFold(memberPerm, permission) {
					return errors.New("member already has the permission")
				}
			}
			team.Members[i].Permissions = append(team.Members[i].Permissions, permission)
			Teams.Set(teamAddress.String(), team)
			return nil
		}
	}

	return errors.New("member not found")
}

// RemoveMemberPermission removes a permission from a specific member of the team
func RemoveMemberPermission(teamAddress std.Address, userAddress std.Address, permission string) error {
	team, err := getTeam(teamAddress)
	if err != nil {
		return err
	}

	// Find the member and remove the permission
	for i, member := range team.Members {
		if member.User.Address == userAddress {
			for j, memberPerm := range member.Permissions {
				if strings.EqualFold(memberPerm, permission) {
					team.Members[i].Permissions = append(member.Permissions[:j], member.Permissions[j+1:]...)
					Teams.Set(teamAddress.String(), team)
					return nil
				}
			}
			return errors.New("permission not found for the member")
		}
	}

	return errors.New("member not found")
}

// helper function to retrieve a team by address
func getTeam(teamAddress std.Address) (*Team, error) {
	team, exists := Teams.Get(teamAddress.String())
	if !exists {
		return nil, errors.New("team not found")
	}
	return team.(*Team), nil
}

// GetTeamByName retrieves a team by its name
func GetTeamByName(teamName string) (*Team, error) {
	teamAddress, exists := TeamNames.Get(strings.ToLower(teamName))
	if !exists {
		return nil, errors.New("team not found")
	}
	return getTeam(teamAddress.(std.Address))
}
