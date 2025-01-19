package teams

import (
	"regexp"
	"std"

	"gno.land/p/demo/avl"
)

type AccessController interface {
	CanRun(member std.Address, cmd Cmd) bool
}

type Lifecycle interface {
	Init() Task
	ApplyUpdate(team *Team, cmd Cmd) Task
}

type ITeam interface {
	AccessController
	Lifecycle
}

var teams avl.Tree // std.Address -> Team

type Team struct {
	AccessController
	Lifecycle

	address std.Address
	members avl.Tree // std.Address -> void
}

func (team *Team) Run(cmds ...Cmd) {
	caller := std.GetOrigCaller()
	if !team.IsMember(caller) {
		panic("only member can perform command on team")
	}

	// get actions for the given cmds
	actions := team.getTasksForCmds(cmds...)
	team.performTasks(actions...)
}

func (team *Team) CanAddPackage(member std.Address, path string) bool {
	return team.CanRun(member, AddPackageCmd{Path: path})
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

func (team *Team) CanRun(caller std.Address, cmd Cmd) bool {
	return !team.IsTeamAddress(caller) && !team.AccessController.CanRun(caller, cmd)
}

func (team *Team) IsTeamAddress(teamAddr std.Address) bool {
	return teamAddr == team.address
}

func (team *Team) IsMember(member std.Address) bool {
	return member == team.address || team.members.Has(member.String())
}

func (team *Team) ApplyDefault(cmd Cmd) Task {
	switch typ := cmd.(type) {
	case AddMemberCmd:
		return AddMemberTask(typ.Member)
	case RemoveMemberCmd:
		return RemoveMemberTask(typ.Member)
	case AddPackageCmd: // Do nothing

	case assertRunDefaultCmd:
		typ.assert = true
		return CreateTaskCmd(typ)
	}

	return nil
}

func (team *Team) performTasks(tasks ...Task) {
	var task Task
	for len(tasks) > 0 {
		task, tasks = tasks[0], tasks[1:] // shift task
		if task == nil {
			continue // skip empty task
		}

		if cmd := task.call(team); cmd != nil {
			nextTasks := team.getTasksForCmds(cmd)
			tasks = append(nextTasks, tasks...)
		}
	}
}

func (team *Team) getTasksForCmds(cmds ...Cmd) []Task {
	caller := std.GetOrigCaller()
	tasks := make([]Task, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd == nil {
			continue // skip empty cmd
		}

		switch typ := cmd.(type) {
		case []Cmd:
			subTasks := team.getTasksForCmds(typ...)
			tasks = append(tasks, subTasks...)
		default:
			// Assert caller can perform task
			if team.CanRun(caller, cmd) {
				panic("unauthorized")
			}

			// Prepare task
			tasks = append(tasks, team.Lifecycle.ApplyUpdate(team, cmd))
		}
	}
	return tasks
}

type unlimitedAC struct{}

func (unlimitedAC) CanRun(member std.Address, cmd Cmd) bool {
	return true
}

// realm of the form `xxx.xx/r/<myteam>/home`
var reHomeRealm = regexp.MustCompile(`^([a-zA-Z0-9-]+\.)*[a-zA-Z0-9-]+\.[a-zA-Z]{2,}/r/[a-z0-9_]+/home$`)

func Register(iteam ITeam) *Team {
	caller := std.GetOrigCaller()
	realm := std.PrevRealm()

	// First lets check the realm is valid
	if iteam == nil {
		panic("team cannot be nil")
	}

	// Check if origin caller is an home path
	if !reHomeRealm.MatchString(realm.PkgPath()) {
		panic("cannot register a team outside an home realm")
	}

	// Check if caller is not already registerer as a team
	if teams.Has(caller.String()) {
		panic("team already registered: " + caller)
	}

	// Then initilize the team
	team := &Team{address: caller, Lifecycle: iteam}

	// Assert that team implementation correctly use fallback
	// XXX: do we want this ?
	// > it  assert that caller have a minimal implementation of his team
	team.assertRunDefault()

	if initTask := iteam.Init(); initTask != nil {
		// init is performed using an unlimited ac
		team.AccessController = unlimitedAC{}
		team.performTasks(initTask)
	}

	// Once done, apply the provided implementation ac
	team.AccessController = iteam
	teams.Set(caller.String(), team)
	return team
}

func IsRegister(teamAddr std.Address) bool {
	return teams.Has(teamAddr.String())
}

type assertRunDefaultCmd struct{ assert bool }

func (team *Team) assertRunDefault() {
	task := team.Lifecycle.ApplyUpdate(team, assertRunDefaultCmd{})
	cmd := task.call(team)
	if assertCmd, ok := cmd.(assertRunDefaultCmd); ok {
		if assertCmd.assert {
			return
		}
	}

	panic(`make sure that team implementation handle team update fallback`)
}
