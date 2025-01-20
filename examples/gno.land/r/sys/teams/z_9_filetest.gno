// PKGPATH: gno.land/r/myteam/home

// This is an example of a more advanced usage of team
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
	return nil
}

func (m *MyTeam) CanRun(member std.Address, cmd teams.Cmd) bool {
	level := m.GetLevel(member)
	shouldBeLevelMinimum := func(target Level) bool {
		return level >= target
	}

	// Team action
	switch typ := cmd.(type) {
	case SetLevelCmd:
		// Can only set level on inferior level
		return level > typ.Level
	case teams.RemoveMemberCmd:
		return level >= Level3
	case teams.AddMemberCmd:
		return level >= Level2
	case teams.AddPackageCmd:
		return level >= Level1
	}

	return false
}

type SetLevelCmd struct {
	Member std.Address
	Level
}

func (SetLevelCmd) Name() string { return "SetLevel" }

func (m *MyTeam) ApplyUpdate(cmd teams.Cmd) teams.Task {
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
	case teams.RemoveMemberCmd:
		return teams.ChainTasks(
			// Add a new member
			teams.RemoveMemberTask(typ.Member),
			// Promote it to level 1
			teams.CreateTaskCmd(SetLevelCmd{
				Member: typ.Member,
				Level:  LevelUnknown,
			}),
		)

	}

	return teams.ApplyDefault(cmd)
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
	// inherit all methods from Team
	myteam.Team = teams.Register(&myteam)
}

func main() {
	println(" -> register myteam")

	myteamUser := myteam.Address()

	// Setup team user address
	println("myteamUser is team address:", myteam.IsTeamAddress(myteamUser))
	println("myteamUser is not member:", !myteam.HasMember(myteamUser))
	println("myteamUser has not level:", myteam.GetLevel(myteamUser) == LevelUnknown)
	println("myteamUser can add package:", myteam.CanAddPackage(myteamUser))

	// Setup test users
	alice := testutils.TestAddress("alice")
	bob := testutils.TestAddress("bob")

	println("alice cannot add a package:", !myteam.CanAddPackage(alice))

	// Register alice to the team
	println(" -> adding alice as a member")
	myteam.AddMember(alice)
	println("alice is member:", myteam.HasMember(alice))
	println("alice is level_1:", myteam.GetLevel(alice) == Level1)
	println("alice can add package:", myteam.CanAddPackage(alice))
	println("bob cannot add package:", !myteam.CanAddPackage(bob))

	// Alice should not be able to add a member on level1
	println("as level_1, alice cannot add bob as member:", !myteam.CanAddMember(alice, bob))

	// Update alice to Level4
	println(" -> setting alice to level 4")
	myteam.SetLevel(alice, Level4)
	println("alice is level_4:", myteam.GetLevel(alice) == Level4)
	println("alice can add bob as member:", myteam.CanAddMember(alice, bob))

	println(" -> burn team address")
	myteam.BurnTeamAddress()
	println("myteamUser is not member:", !myteam.HasMember(myteamUser))
	println("myteamUser is level_Unknown:", myteam.GetLevel(myteamUser) == LevelUnknown)
	println("myteamUser cannot add package:", !myteam.CanAddPackage(myteamUser))

	println("alice is still level_4:", myteam.GetLevel(alice) == Level4)
	println("alice can still add bob as member:", myteam.CanAddMember(alice, bob))

	// Set caller to alice
	println(" -> setting alice as origin caller")
	std.TestSetOrigCaller(alice)

	// alice add member bob
	println(" -> alice add bob as member")
	myteam.AddMember(bob)

	println("bob is member:", myteam.HasMember(bob))
	println("bob is level_1:", myteam.GetLevel(bob) == Level1)
	println("bob can add package:", myteam.CanAddPackage(bob))

	println(" -> removing bob")
	myteam.RemoveMember(bob)
	println("bob is not member:", !myteam.HasMember(bob))

}

// Output:
//  -> register myteam
// myteamUser is team address: true
// myteamUser is not member: true
// myteamUser has not level: true
// myteamUser can add package: true
// alice cannot add a package: true
//  -> adding alice as a member
// alice is member: true
// alice is level_1: true
// alice can add package: true
// bob cannot add package: true
// as level_1, alice cannot add bob as member: true
//  -> setting alice to level 4
// alice is level_4: true
// alice can add bob as member: true
//  -> burn team address
// myteamUser is not member: true
// myteamUser is level_Unknown: true
// myteamUser cannot add package: true
// alice is still level_4: true
// alice can still add bob as member: true
//  -> setting alice as origin caller
//  -> alice add bob as member
// bob is member: true
// bob is level_1: true
// bob can add package: true
//  -> removing bob
// bob is not member: true
