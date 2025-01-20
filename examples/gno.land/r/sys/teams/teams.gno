package teams

import (
	"regexp"
	"std"

	"gno.land/p/demo/avl"
)

// AccessController defines the interface for controlling access to team commands.
// Implementations should define the logic to determine if a member can run a specific command.
type AccessController interface {
	// CanRun determines if a member is authorized to execute a specific command.
	CanRun(member std.Address, cmd Cmd) bool
}

// Lifecycle defines the interface for managing the lifecycle of a team.
// It provides methods for initializing a team and applying updates through commands.
// Implementations of this interface should define how a team is set up initially
// and how it responds to changes over time.
type Lifecycle interface {
	// Init initializes the team with a series of commands.
	// It returns a Task that encapsulates the initialization logic.
	//
	// The Init method is called during the creation of a team to perform
	// initial setup actions such as adding founding members or configuring
	// initial settings. Notably, this method operates with elevated privileges,
	// meaning it does not enforce access control checks via the AccessController.
	// This allows the initial setup to bypass usual restrictions, enabling
	// foundational configuration without member-specific constraints.
	//
	// Example:
	//   func (m *MyTeam) Init() teams.Task {
	//       return teams.CreateTaskCmd(
	//           teams.AddMemberCmd{myteam.MyUser}, // Add MyUser as a member
	//           SetLevelCmd{myteam.MyUser, Level4}, // Set MyUser's level to Level4
	//       )
	//   }
	Init() Task

	// ApplyUpdate applies an update to the team based on a given command.
	// It returns a Task that represents the execution of this update.
	//
	// The ApplyUpdate method is used to modify the team's state in response
	// to commands such as adding or removing members, or changing member roles.
	// Unlike Init, this method enforces access control checks using the
	// AccessController. Before executing a command, ApplyUpdate ensures that
	// the member issuing the command has the necessary permissions, maintaining
	// the security and integrity of team operations.
	//
	// Example:
	//   func (m *MyTeam) ApplyUpdate(cmd teams.Cmd) teams.Task {
	//       switch typ := cmd.(type) {
	//       case SetLevelCmd:
	//           return teams.CreateTask(func(_ *teams.Team) teams.Cmd {
	//               m.levels.Set(typ.Member.String(), typ.Level)
	//               return nil
	//           })
	//       }
	//       return team.ApplyDefault(cmd)
	//   }
	ApplyUpdate(cmd Cmd) Task
}

// ITeam combines the AccessController and Lifecycle interfaces to define a complete
// team interface. It ensures that a team has both access control and lifecycle
// management capabilities.
type ITeam interface {
	AccessController
	Lifecycle
}

var teams avl.Tree // std.Address ->  *Team

type Team struct {
	AccessController
	Lifecycle

	burned  bool
	address std.Address // Origin address of the team
	members avl.Tree    // std.Address -> void

	// internal
	isUpdating bool
}

// Realm of the form `xxx.xx/r/<myteam>/home`
var reHomeRealm = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}/r/[a-z0-9_]+/home$`)

// Register creates and registers a new team in the `r/sys/teams` realm, allowing a registered
// user to transform into a team by publishing a contract that registers members in a registry.
//
// The `Register` function verifies that the caller is a valid home path (`r/<handle>/home`),
// using `<handle>` as the team name, and ensures that the caller is not already registered as a team.
func Register(iteam ITeam) *Team {
	caller := std.GetOrigCaller()
	realm := std.PrevRealm()

	// Check if caller is not already registered as a team
	if teams.Has(caller.String()) {
		panic("team already registered: " + caller)
	}

	// Check if origin caller is a home path
	if !reHomeRealm.MatchString(realm.PkgPath()) {
		panic("cannot register a team outside a home realm")
	}

	// Then initilize the team
	team := &Team{
		address:          caller,
		Lifecycle:        iteam,
		AccessController: iteam,
	}

	// Assert that team implementation correctly uses fallback
	// XXX: do we want this?
	// It asserts that caller has a minimal implementation of its team
	assertDefaultUpdate(team)

	if initTask := team.Init(); initTask != nil {
		team.performTasks(caller, initTask)
	}

	// All set, register the team
	teams.Set(caller.String(), team)
	return team
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
	if !team.burned && team.IsTeamAddress(member) {
		return true
	}

	return team.AccessController != nil && team.AccessController.CanRun(member, cmd)
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

func (team *Team) CanAddPackage(member std.Address, path string) bool {
	return team.CanRun(member, AddPackageCmd{Path: path})
}

// BurnTeamAddress prevent the team address from managing the team, leaving it to the members.
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

func ApplyDefault(cmd Cmd) Task {
	switch typ := cmd.(type) {
	case assertDefaultUpdateCmd:
		typ.assert = true
		return CreateTaskCmd(typ)
	case AddMemberCmd:
		return AddMemberTask(typ.Member)
	case RemoveMemberCmd:
		return RemoveMemberTask(typ.Member)
	case AddPackageCmd: // XXX:
		return nil
	case UpdateAccessControllerCmd:
		// this commands are not supported by default
	}

	panic("command not supported")
}

func IsRegister(teamAddr std.Address) bool {
	return teams.Has(teamAddr.String())
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
				panic("unauthorized command: " + "[" + cmd.Name() + "]")
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
