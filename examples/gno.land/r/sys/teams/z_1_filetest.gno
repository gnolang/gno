// PKGPATH: gno.land/r/myteam/home

// This the most minimal usage of team
package home

import (
	"gno.land/p/demo/testutils"
	"gno.land/r/sys/teams"
)

var myteam *teams.Team

// var myteam *Team
func init() {
	myteam = teams.Register(nil)
}

func main() {
	alice := testutils.TestAddress("alice")
	teamAddress := myteam.Address()

	// Setup team user address
	println("team address is team address:", myteam.IsTeamAddress(teamAddress))
	println("team address is not member:", !myteam.HasMember(teamAddress))
	println("team address can add package:", myteam.CanAddPackage(teamAddress, ""))
	println("team address can add member alice:", myteam.CanAddMember(teamAddress, alice))

	println("alice is not member:", !myteam.HasMember(alice))
	println(" -> adding alice as member")
	myteam.AddMember(alice)
	println("alice is now a member:", myteam.HasMember(alice))
}
