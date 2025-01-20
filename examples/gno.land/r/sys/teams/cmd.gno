package teams

import (
	"errors"
	"std"
	"strings"
)

var ErrAlreadyExist = errors.New("already exists")
var ErrDoesNotExist = errors.New("does not exist")

// Cmd represents a command that can be executed on a team.
// Each command should implement the Name method to provide a string identifier.
type Cmd interface {
	Name() string
}

type Cmds []Cmd

func (cmds Cmds) Name() string {
	var str strings.Builder
	str.WriteRune('[')
	for i, cmd := range cmds {
		if i > 0 {
			str.WriteRune(',')
		}

		str.WriteString(cmd.Name())
	}
	str.WriteRune(']')
	return str.String()
}

// AddMemberCmd represents a command to add a member to the team.
type AddMemberCmd struct {
	Member std.Address
}

func (cmd AddMemberCmd) Name() string { return "AddMember" }

// AddMemberTask creates a task to add a member to the team.
func AddMemberTask(member std.Address) Task {
	return CreateTask(func(t *Team) Cmd {
		// Cannot add team address as a member
		if t.IsTeamAddress(member) {
			panic("cannot add team address as a member")
		}

		if t.members.Has(member.String()) {
			panic(ErrAlreadyExist)
		}
		t.members.Set(member.String(), struct{}{})
		return nil
	})
}

// RemoveMemberCmd represents a command to remove a member from the team.
type RemoveMemberCmd struct {
	Member std.Address
}

func (cmd RemoveMemberCmd) Name() string { return "RemoveMember" }

// RemoveMemberTask creates a task to remove a member from the team.
func RemoveMemberTask(member std.Address) Task {
	return CreateTask(func(t *Team) Cmd {
		if !t.members.Has(member.String()) {
			panic(ErrDoesNotExist)
		}
		t.members.Remove(member.String())
		return nil
	})
}

// The command bellow should be use with precaution

// UpdateAccessControllerCmd represents a command to update the team's access controller.
type UpdateAccessControllerCmd struct {
	AccessController
}

func (cmd UpdateAccessControllerCmd) Name() string { return "UpdateAccessController" }

// UpdateAccessControllerTask creates a task to update the team's access controller.
func UpdateAccessControllerTask(ac AccessController) Task {
	return CreateTask(func(t *Team) Cmd {
		t.AccessController = ac
		return nil
	})
}

// UpdateLifecycleCmd represents a command to update the team's access controller.
type UpdateLifecycleCmd struct {
	Lifecycle
}

func (cmd UpdateLifecycleCmd) Name() string { return "UpdateLifecycle" }

// UpdateLifecycleTask creates a task to update the team's access controller.
func UpdateLifecycleTask(ac Lifecycle) Task {
	return CreateTask(func(t *Team) Cmd {
		t.Lifecycle = ac
		return nil
	})
}

// BurnTeamAddressCmd represents a command to burn the team's address.
type BurnTeamAddressCmd struct{}

func (cmd BurnTeamAddressCmd) Name() string { return "BurnTeamAddress" }

var BurnTeamAddressTask = CreateTaskCmd(BurnTeamAddressCmd{})

// AddPackageCmd represents a command to add a package to the team.
type AddPackageCmd struct {
	Path string
}

func (cmd AddPackageCmd) Name() string { return "AddPackage" }
