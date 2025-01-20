package teams

import (
	"std"

	"gno.land/p/demo/avl"
)

type Team struct {
	AccessController
	Lifecycle

	address std.Address // Origin address of the team
	members avl.Tree    // std.Address -> void
	burned  bool

	// internal
	isUpdating bool
}

// Run executes a series of commands on the team, ensuring that only authorized
// members can perform these operations. It translates commands into tasks and
// executes them to update the team's state.
//
// This is the entrypoint to execute command on the team.
//
// Flow:
//
//  1. **Caller Verification**: The method retrieves the original caller's address
//     and checks if they are a member of the team. If not, it panics, preventing
//     unauthorized access.
//
//  2. **Command Translation**: The method translates the provided commands into
//     tasks. This involves checking if each command is authorized for execution
//     by the caller and preparing the corresponding tasks.
//
//  3. **Task Execution**: The method executes the tasks, ensuring that each
//     task is executed in sequence, effectively updating the team's state.
func (team *Team) Run(cmds ...Cmd) {
	if !team.IsRegister() {
		panic("team is not register")
	}

	caller := std.GetOrigCaller()
	if !team.IsTeamAddress(caller) && !team.HasMember(caller) {
		panic("only members / team address can perform commands on the team")
	}

	// Get tasks for the given cmds
	tasks := team.getTasksForCmds(caller, cmds...)
	// Perform tasks
	team.performTasks(caller, tasks...)
}

// Default Access Controller

func (team *Team) CanRun(member std.Address, cmd Cmd) bool {
	if !team.IsRegister() { // team havn't been registered
		return false
	}

	isTeamAddress := team.IsTeamAddress(member)
	isMember := team.HasMember(member)

	if !isMember && !isTeamAddress {
		return false
	}

	// TeamAddress has all the rights, until it has been burned.
	if !team.burned && isTeamAddress {
		return true
	}

	// If access controller has been set, check access using `CanRun`.
	if team.AccessController != nil {
		return team.AccessController.CanRun(member, cmd)
	}

	// Fallback on Default Permission
	switch cmd.(type) {
	case AddMemberCmd, AddPackageCmd:
		return true // any member can do it
	default:
	}

	return false
}

// Default Lifecycle implementation

func (team *Team) Init() Task {
	if team.Lifecycle != nil {
		return team.Lifecycle.Init()
	}

	return nil
}

func (team *Team) ApplyUpdate(cmd Cmd) Task {
	if team.Lifecycle != nil {
		return team.Lifecycle.ApplyUpdate(cmd)
	}

	return ApplyDefault(cmd)
}

func ApplyDefault(cmd Cmd) Task {
	switch typ := cmd.(type) {
	case assertDefaultUpdateCmd:
		typ.assert = true
		return CreateTaskCmd(typ) // send it back
	case AddMemberCmd:
		return AddMemberTask(typ.Member)
	case RemoveMemberCmd:
		return RemoveMemberTask(typ.Member)
	case AddPackageCmd: // XXX:
		return nil
	case UpdateAccessControllerCmd:
		// Default UpdateAccessController can only update an empty
		// AccessController
		return CreateTask(func(t *Team) Cmd {
			if t.AccessController != nil {
				panic(`AccessController already set`)
			}

			t.AccessController = typ.AccessController
			return nil
		})
	case UpdateLifecycleCmd:
		// Default UpdateLifecycle can only update an empty
		// Lifecycle
		return CreateTask(func(t *Team) Cmd {
			if t.Lifecycle != nil {
				panic(`lifecycle already set`)
			}

			t.Lifecycle = typ.Lifecycle
			return nil
		})
	}

	panic("command not supported: [" + cmd.Name() + `]`)
}

func (team *Team) HasMember(member std.Address) bool {
	return team.members.Has(member.String())
}

func (team *Team) IsTeamAddress(teamAddr std.Address) bool {
	return teamAddr == team.address
}

func (team *Team) IsRegister() bool {
	return team.address != ""
}

func (team *Team) Address() std.Address {
	return team.address
}

// Shortcut command

func (team *Team) CanAddPackage(member std.Address) bool {
	return team.CanRun(member, AddPackageCmd{})
}

// BurnTeamAddress prevent the team address from managing the team, leaving it
// to the members.
// BurnTeamAddress is a special command, and cannot be handle by team implementation.
// WARNING: This is irreversible
func (team *Team) BurnTeamAddress() {
	team.Run(BurnTeamAddressCmd{})
}

func (team *Team) AddMember(member std.Address) {
	team.Run(AddMemberCmd{Member: member})
}

func (team *Team) CanAddMember(member, target std.Address) bool {
	return team.CanRun(member, AddMemberCmd{Member: member})
}

func (team *Team) RemoveMember(member std.Address) {
	team.Run(RemoveMemberCmd{Member: member})
}

func (team *Team) CanRemoveMember(member, target std.Address) bool {
	return team.CanRun(member, RemoveMemberCmd{Member: target})
}

func (team *Team) performTasks(caller std.Address, tasks ...Task) {
	if team.isUpdating {
		panic(`cannot perform task while updating, ensure returning the task instead`)
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
		case BurnTeamAddressCmd: // special case, can't be ignored
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

			// XXX: do we fallthrough here? can be usefull for
			// people to take some action but can also let people
			// prevent burning by using `panic`
		default:
			// Assert caller can perform task
			if !team.CanRun(caller, cmd) {
				panic("unauthorized command for caller: [" + cmd.Name() + "]")
			}

			// Prepare task
			tasks = append(tasks, team.ApplyUpdate(cmd))
		}
	}
	return tasks
}

type assertDefaultUpdateCmd struct{ assert bool }

func (assertDefaultUpdateCmd) Name() string { return "assertDefaultUpdateCmd" }

func assertDefaultUpdate(team *Team) {
	task := team.ApplyUpdate(assertDefaultUpdateCmd{})
	cmd := task.call(team)
	if assertCmd, ok := cmd.(assertDefaultUpdateCmd); ok {
		if assertCmd.assert {
			return
		}
	}

	panic("ensure that team implementation handles team update fallback")
}
