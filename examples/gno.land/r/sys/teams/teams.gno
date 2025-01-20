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
	// initial settings.
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
	// Before executing a command, ApplyUpdate ensures that the member
	// issuing the command has the necessary permissions, maintaining the
	// security and integrity of team operations.
	//
	// ApplyUpdate should always fallback on `teams.ApplyDefault`.
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

func Get(teamAddr std.Address) *Team {
	if t, ok := teams.Get(teamAddr.String()); ok {
		return t.(*Team)
	}

	return nil
}

func IsRegister(teamAddr std.Address) bool {
	return teams.Has(teamAddr.String())
}
