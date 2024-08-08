package teams

import (
	"errors"
	"std"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/users"
)

//----------------------------------------
// Types

type Team struct {
	Address     std.Address
	TeamName    string
	TeamProfile string
	Creator     std.Address
}

var (
	Teams    avl.Tree // t.Address -> t.TeamName:t.TeamProfile
	Members  avl.Tree // t.Address -> []*users.User
	Creators avl.Tree // t.Creator -> t.Address
)

type TeamManager struct {
	Teams    *avl.Tree
	Members  *avl.Tree
	Creators *avl.Tree
}

func NewTeamManager() *TeamManager {
	tm := &TeamManager{
		Teams:    avl.NewTree(),
		Members:  avl.NewTree(),
		Creators: avl.NewTree(),
	}
	return tm
}

func (tm *TeamManager) AddTeam(t Team) error {
	// Check if team already exists
	if tm.Teams.Size() != 0 {
		_, exist := tm.Teams.Get(t.Address.String())
		if exist {
			return ErrTeamAddressAlreadyExists
		}
	}

	// Add the team
	tm.Teams.Set(t.Address.String(), t.TeamName+":"+t.TeamProfile)
	tm.Creators.Set(t.Creator.String(), t.Address.String())
	return nil
}

func (tm *TeamManager) AddUserToTeam(teamAddress std.Address, user users.User) error {
	// Check if the team exists
	_, teamExists := tm.Teams.Get(teamAddress.String())
	if !teamExists {
		return ErrTeamNotFound
	}

	// Get current members
	members, _ := tm.Members.Get(teamAddress.String())
	var memberList []*users.User
	if members != nil {
		memberList = members.([]*users.User)
	} else {
		memberList = []*users.User{}
	}

	// Check if user is already a member
	for _, member := range memberList {
		if member.Address == user.Address {
			return ErrUserAlreadyMember
		}
	}

	// Add user to team
	memberList = append(memberList, &user)
	tm.Members.Set(teamAddress.String(), memberList)
	return nil
}

func (tm *TeamManager) RemoveUserFromTeam(teamAddress std.Address, userAddress std.Address) error {
	// Check if the team exists
	_, teamExists := tm.Teams.Get(teamAddress.String())
	if !teamExists {
		return ErrTeamNotFound
	}

	// Get current members
	members, _ := tm.Members.Get(teamAddress.String())
	var memberList []*users.User
	if members != nil {
		memberList = members.([]*users.User)
	} else {
		return ErrUserNotMember
	}

	// Check if user is a member
	userFound := false
	for i, member := range memberList {
		if member.Address == userAddress {
			// Remove user from team
			memberList = append(memberList[:i], memberList[i+1:]...)
			userFound = true
			break
		}
	}

	if !userFound {
		return ErrUserNotMember
	}

	// Update members list
	tm.Members.Set(teamAddress.String(), memberList)
	return nil
}

func (tm *TeamManager) GetAllUsersInTeam(teamAddress std.Address) ([]users.User, error) {
	// Check if the team exists
	_, teamExists := tm.Teams.Get(teamAddress.String())
	if !teamExists {
		return nil, ErrTeamNotFound
	}

	// Get current members
	members, _ := tm.Members.Get(teamAddress.String())
	var memberList []*users.User
	if members != nil {
		memberList = members.([]*users.User)
	} else {
		return nil, ErrUserNotMember
	}

	// Convert []*users.User to []users.User
	userList := make([]users.User, len(memberList))
	for i, member := range memberList {
		userList[i] = *member
	}

	return userList, nil
}

func (t *Team) Render(tm *TeamManager) string {
	str := "## team " + t.TeamName + "\n" +
		"\n" +
		" * address = " + t.Address.String() + "\n" +
		" * creator = " + t.Creator.String() + "\n" +
		"\n" +
		t.TeamProfile + "\n" +
		"\n" +
		"### Members\n"

	// Get the members of the team
	members, _ := tm.GetAllUsersInTeam(t.Address)
	if members != nil {
		for _, member := range members {
			str += " * " + member.Name + " (" + member.Address.String() + "): " + member.Profile + "\n"
		}
	} else {
		str += "No members found.\n"
	}

	return str
}
