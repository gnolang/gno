// PKGPATH: gno.land/r/myteam/home
package home

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/r/sys/teams"
)

type Level int

const (
	LevelUnknown Level = iota // lowest
	Level1
	Level2
	Level3
	Level4
)

const (
	Promote teams.Action = "promote"
	Demote               = "demote"
)

var teamAddress std.Address

var MyTeam team.Team

type MyTeam struct {
	*teams.Team
	levels avl.Tree // std.Address -> Level
}

func (m *MyTeam) CanPerform(member std.Address, action teams.Action, ressource interface{}) bool {
	var level Level
	if mLevel, ok := m.levels.Get(member); ok {
		level = mLevel
	}

	// Base action
	switch do {
	case teams.Add:
		if level >= Level3 {
			return true
		}

	case teams.Delete:
		if level >= Level4 {
			return true
		}

	case teams.AddPkg:
		if level >= Level1 {
			return true
		}
	}

	return false // noop
}

type PromoteAction struct {
	Level
	Member std.Address
}

func (m *MyTeam) Promote(member std.Address, toLevel Level) {
	m.Perform(Promote, PomoteAction{
		Member: member,
		Level:  toLevel,
	})
}

func init() {
	teamAddress = std.GetOrigCaller()
	myteam.Team = teams.Register(&myteam, func(t *Team) {

		t.RegisterAction
	})
}

func main() {
	println(teams.IsRegister(teamAddress))
}
