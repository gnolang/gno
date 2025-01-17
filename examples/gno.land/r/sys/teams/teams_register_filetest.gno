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

var baseteam teams.BaseTeam
var teamAddress std.Address

type Action teams.Action

var team *Team

type MyTeam struct {
	levels avl.Tree // std.Address -> Level
}

func (m *MyTeam) Can(member std.Address, do teams.Action, args ...string) bool {
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

func (m *MyTeam) Promote(by std.Address, on ...std.Address) {

}

func init() {
	teamAddress = std.GetOrigCaller()

	var myteam MyTeam
	team = teams.Register(&myteam)
}

func main() {
	println(teams.IsRegister(teamAddress))
}
