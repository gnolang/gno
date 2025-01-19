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

func (m *MyTeam) Init() teams.Action {
	caller := std.GetOrigCaller()
	return teams.ActionableMsg(
		// Add caller as member
		teams.AddMemberMsg{caller},
		// Promote caller to level4
		SetLevelMsg{caller, Level4},
	)
}

func (m *MyTeam) CanPerform(member std.Address, msg teams.Msg) bool {
	var level Level
	if mLevel, ok := m.levels.Get(member.String()); ok {
		level = mLevel.(Level)
	}

	shouldLevelMinimum := func(target Level) bool {
		return level >= target
	}

	// Base action
	switch msg.(type) {
	case SetLevelMsg:
		return shouldLevelMinimum(Level4)
	case teams.RemoveMemberMsg:
		return shouldLevelMinimum(Level3)
	case teams.AddMemberMsg:
		return shouldLevelMinimum(Level2)
	case teams.AddPackageMsg:
		return shouldLevelMinimum(Level1)
	}

	return false
}

func (m *MyTeam) Update(team *teams.Team, msg teams.Msg) teams.Action {
	switch typ := msg.(type) {
	case SetLevelMsg:
		return teams.Actionable(func(_ *teams.Team) teams.Msg {
			mkey := typ.Member.String()
			m.levels.Set(mkey, typ.Level)
			return nil
		})
	case teams.AddMemberMsg:
		return teams.ChainActions(
			// Add a new member
			teams.AddMemberAction(typ.Member),
			// Promote it to level 1
			teams.ActionableMsg(SetLevelMsg{
				Member: typ.Member,
				Level:  Level1,
			}),
		)
	}

	return team.PerformDefault(msg)
}

type SetLevelMsg struct {
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
	m.Team.Perform(SetLevelMsg{
		Member: member,
		Level:  level,
	})
}

var myteam MyTeam

func init() {
	myteam.Team = teams.Register(&myteam)
}

func main() {
	// Setup user for test
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	println("myteam can add a package:", myteam.CanPerform(alice, teams.AddPackageMsg{}))

	// Register alice to the team
	myteam.AddMember(alice)
	println("alice is member: ", myteam.HasMember(alice))
	println("alice is level 1: ", myteam.GetLevel(alice) == Level1)

	// Should be able to add a package on level1
	println("alice can add package:", myteam.CanPerform(alice, teams.AddPackageMsg{}))
	// Should be able to add a package on level1
	println("bob cannot add package:", myteam.CanPerform(bob, teams.AddPackageMsg{}))

	// Should not be able to add a member on level1
	println("alice cannot add bob as member:",
		myteam.CanPerform(alice, teams.AddMemberMsg{
			Member: bob,
		}))

	// Update alice to Level4
	myteam.SetLevel(alice, Level4)
	println("alice is level 4: ", myteam.GetLevel(alice) == Level4)
	println("alice can add bob as member:",
		myteam.CanPerform(alice, teams.AddMemberMsg{
			Member: bob,
		}))

	// Set caller to alice
	std.TestSetOrigCaller(alice)

	// alice add member bob
	println("adding bob")
	myteam.AddMember(bob)
	println("bob is member: ", myteam.HasMember(alice))
	println("bob is level 1: ", myteam.GetLevel(alice) == Level1)

	// Check if alice can perform add package
	// println(teams.CanPerform(myTeamUser, alice, teams.AddPackageMsg{}))

	// Set alice as caller
	// std.TestSetOrigCaller(alice)
	// Try to perform add package

	// println(teams.CanPerform(myTeamUser, alice, teams.AddPackageMsg{}))

	// bob := testutils.TestAddress("bob")

	// println(teams.IsRegister(myteam.address))
	// myteam.address = teams.Register(&myteam)
	// println(teams.IsRegister(myteam.address))

}
