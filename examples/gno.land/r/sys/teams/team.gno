package teams

import (
	"std"

	"gno.land/p/demo/avl"
)

type Team struct {
	AccessController
	Lifecycle

	address    std.Address // Origin address of the team
	members    avl.Tree    // std.Address -> void
	burned     bool
	isUpdating bool // internal flag for update status
}

// Run executes a series of commands on the team, ensuring that only authorized
// members can perform these operations. It translates commands into tasks and
// executes them to update the team's state.
//
// Flow:
//  1. **Caller Verification**: Checks if the caller is a team member or the team address.
//  2. **Command Translation**: Translates commands into tasks for execution.
//  3. **Task Execution**: Executes tasks in sequence, updating the team's state.
func (team *Team) Run(cmds ...Cmd) {
	if !team.IsRegistered() {
		panic("team is not registered")
	}

	caller := std.GetOrigCaller()
	if !team.IsTeamAddress(caller) && !team.HasMember(caller) {
		panic("only members or team address can perform commands on the team")
	}

	// Get tasks for the given cmds
	tasks := team.getTasksForCmds(caller, cmds...)
	// Perform tasks
	team.performTasks(caller, tasks...)
}

// CanRun checks if a member can run a specific command.
func (team *Team) CanRun(member std.Address, cmd Cmd) bool {
	if !team.IsRegistered() { // team hasn't been registered
		return false
	}

	isTeamAddress := team.IsTeamAddress(member)
	isMember := team.HasMember(member)

	if !isMember && !isTeamAddress {
		return false
	}

	// TeamAddress has all the rights until it has been burned.
	if !team.burned && isTeamAddress {
		return true
	}

	// If an AccessController is set, delegate the check.
	if team.AccessController != nil {
		return team.AccessController.CanRun(member, cmd)
	}

	// Fallback on Default Permission
	switch cmd.(type) {
	case AddMemberCmd, AddPackageCmd:
		return true // any member can do it
	default:
		return false
	}
}

// Init initializes the team lifecycle.
func (team *Team) Init() Task {
	if team.Lifecycle != nil {
		return team.Lifecycle.Init()
	}
	return nil
}

// ApplyUpdate applies a command update to the team.
func (team *Team) ApplyUpdate(cmd Cmd) Task {
	if team.Lifecycle != nil {
		return team.Lifecycle.ApplyUpdate(cmd)
	}
	return ApplyDefault(cmd)
}

// ApplyDefault handles default command updates.
func ApplyDefault(cmd Cmd) Task {
	switch typ := cmd.(type) {
	case assertDefaultUpdateCmd:
		typ.assert = true
		return CreateTaskCmd(typ) // send it back
	case AddMemberCmd:
		return AddMemberTask(typ.Member)
	case RemoveMemberCmd:
		return RemoveMemberTask(typ.Member)
	case AddPackageCmd: // XXX: Consider implementation
		return nil
	case UpdateAccessControllerCmd:
		return CreateTask(func(t *Team) Cmd {
			if t.AccessController != nil {
				panic("AccessController already set")
			}
			t.AccessController = typ.AccessController
			return nil
		})
	case UpdateLifecycleCmd:
		return CreateTask(func(t *Team) Cmd {
			if t.Lifecycle != nil {
				panic("lifecycle already set")
			}
			t.Lifecycle = typ.Lifecycle
			return nil
		})
	default:
		panic("command not supported: [" + cmd.Name() + "]")
	}
}

// HasMember checks if a given address is a member of the team.
func (team *Team) HasMember(member std.Address) bool {
	return team.members.Has(member.String())
}

// IsTeamAddress checks if a given address is the team's address.
func (team *Team) IsTeamAddress(teamAddr std.Address) bool {
	return teamAddr == team.address
}

// IsRegistered checks if the team is registered.
func (team *Team) IsRegistered() bool {
	return team.address != ""
}

// Address returns the team's address.
func (team *Team) Address() std.Address {
	return team.address
}

// CanAddPackage checks if a member can add a package.
func (team *Team) CanAddPackage(member std.Address) bool {
	return team.CanRun(member, AddPackageCmd{})
}

// BurnTeamAddress prevents the team address from managing the team, leaving it to the members.
// WARNING: This is irreversible.
func (team *Team) BurnTeamAddress() {
	team.Run(BurnTeamAddressCmd{})
}

// AddMember adds a member to the team.
func (team *Team) AddMember(member std.Address) {
	team.Run(AddMemberCmd{Member: member})
}

// CanAddMember checks if a member can add another member.
func (team *Team) CanAddMember(member, target std.Address) bool {
	return team.CanRun(member, AddMemberCmd{Member: member})
}

// RemoveMember removes a member from the team.
func (team *Team) RemoveMember(member std.Address) {
	team.Run(RemoveMemberCmd{Member: member})
}

// CanRemoveMember checks if a member can remove another member.
func (team *Team) CanRemoveMember(member, target std.Address) bool {
	return team.CanRun(member, RemoveMemberCmd{Member: target})
}

// performTasks executes a list of tasks.
func (team *Team) performTasks(caller std.Address, tasks ...Task) {
	if team.isUpdating {
		panic("cannot perform task while updating, ensure returning the task instead")
	}

	var task Task
	for len(tasks) > 0 {
		task, tasks = tasks[0], tasks[1:] // Shift task
		if task == nil {
			continue // Skip empty task
		}

		if cmd := task.call(team); cmd != nil {
			nextTasks := team.getTasksForCmds(caller, cmd)
			tasks = append(nextTasks, tasks...)
		}
	}
}

// getTasksForCmds translates commands into tasks.
func (team *Team) getTasksForCmds(caller std.Address, cmds ...Cmd) []Task {
	team.isUpdating = true
	defer func() { team.isUpdating = false }()

	tasks := make([]Task, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd == nil {
			continue // Skip empty cmd
		}

		switch typ := cmd.(type) {
		case Cmds:
			subTasks := team.getTasksForCmds(caller, typ...)
			tasks = append(tasks, subTasks...)
		case BurnTeamAddressCmd:
			if team.burned {
				panic("already burned")
			}

			if !team.IsTeamAddress(caller) {
				panic("only team address can burn")
			}

			tasks = append(tasks, CreateTask(func(t *Team) Cmd {
				team.burned = true
				return nil
			}))

			// XXX: Consider if fallthrough is needed
		default:
			if !team.CanRun(caller, cmd) {
				panic("unauthorized command for caller: [" + cmd.Name() + "]")
			}

			tasks = append(tasks, team.ApplyUpdate(cmd))
		}
	}
	return tasks
}

type assertDefaultUpdateCmd struct{ assert bool }

func (assertDefaultUpdateCmd) Name() string { return "assertDefaultUpdateCmd" }

// assertDefaultUpdate ensures the team implementation handles team update fallback.
func assertDefaultUpdate(team *Team) {
	task := team.ApplyUpdate(assertDefaultUpdateCmd{})
	cmd := task.call(team)
	if assertCmd, ok := cmd.(assertDefaultUpdateCmd); ok && assertCmd.assert {
		return
	}
	panic("ensure that team implementation handles team update fallback")
}
