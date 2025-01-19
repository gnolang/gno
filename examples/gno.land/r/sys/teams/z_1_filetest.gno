// PKGPATH: gno.land/r/myteam/home
package home

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/r/sys/teams"
)

// var myteam *Team

type Level int

const (
	LevelUnknown Level = iota // lowest
	Level1
	Level2
	Level3
	Level4
)

type MyTeam struct {
	*teams.Team
	address std.Address
	levels  avl.Tree // std.Address -> Level
}

func (m *MyTeam) Init() teams.Task {
	caller := std.GetOrigCaller()
	return teams.CreateTaskCmd(
		// Add caller as member
		teams.AddMemberCmd{caller},
		// Promote caller to level4
		SetLevelCmd{caller, Level4},
	)
}

func (m *MyTeam) CanRun(member std.Address, cmd teams.Cmd) bool {
	var level Level
	if mLevel, ok := m.levels.Get(member.String()); ok {
		level = mLevel.(Level)
	}

	shouldLevelMinimum := func(target Level) bool {
		return level >= target
	}

	// Team action
	switch cmd.(type) {
	case SetLevelCmd:
		return shouldLevelMinimum(Level4)
	case teams.RemoveMemberCmd:
		return shouldLevelMinimum(Level3)
	case teams.AddMemberCmd:
		return shouldLevelMinimum(Level2)
	case teams.AddPackageCmd:
		return shouldLevelMinimum(Level1)
	}

	return false
}

func (m *MyTeam) ApplyUpdate(team *teams.Team, cmd teams.Cmd) teams.Task {
	switch typ := cmd.(type) {
	case SetLevelCmd:
		return teams.CreateTask(func(_ *teams.Team) teams.Cmd {
			mkey := typ.Member.String()
			m.levels.Set(mkey, typ.Level)
			return nil
		})
	case teams.AddMemberCmd:
		return teams.ChainTasks(
			// Add a new member
			teams.AddMemberTask(typ.Member),
			// Promote it to level 1
			teams.CreateTaskCmd(SetLevelCmd{
				Member: typ.Member,
				Level:  Level1,
			}),
		)
	}

	return team.ApplyDefault(cmd)
}

type SetLevelCmd struct {
	Member std.Address
	Level
}

func (m *MyTeam) GetLevel(member std.Address) Level {
	if level, ok := m.levels.Get(member.String()); ok {
		return level.(Level)
	}
	return LevelUnknown
}

func (m *MyTeam) SetLevel(member std.Address, level Level) {
	m.Team.Run(SetLevelCmd{
		Member: member,
		Level:  level,
	})
}

var myteam MyTeam

func init() {
	myteam.Team = teams.Register(&myteam)
}

func main() {
	// // Setup team address
	// alice := testutils.TestAddress("alice")

	// Setup user for test
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	println("myteam can add a package:", myteam.CanAddPackage(alice, ""))

	// Register alice to the team
	myteam.AddMember(alice)
	println("alice is member:", myteam.HasMember(alice))
	println("alice is level_1:", myteam.GetLevel(alice) == Level1)
	println("alice can add package:", myteam.CanAddPackage(alice, ""))
	println("bob cannot add package:", myteam.CanAddPackage(bob, ""))

	// Alice should not be able to add a member on level1
	println("as level_1, alice cannot add bob as member:", myteam.CanAddMember(alice, bob))

	// Update alice to Level4
	myteam.SetLevel(alice, Level4)
	println("alice is level_4:", myteam.GetLevel(alice) == Level4)
	println("alice can add bob as member:", myteam.CanAddMember(alice, bob))

	// Set caller to alice
	std.TestSetOrigCaller(alice)

	// alice add member bob
	println("alice add bob as member")
	myteam.AddMember(bob)

	println("bob is member:", myteam.HasMember(alice))
	println("bob is level_1:", myteam.GetLevel(alice) == Level1)

}
