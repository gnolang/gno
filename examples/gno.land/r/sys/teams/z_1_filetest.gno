// PKGPATH: gno.land/r/myteam/home

// This the most minimal usage of team
package home

import (
	"std"

	"gno.land/p/demo/testutils"
	"gno.land/r/sys/teams"
)

func main() {
	myTeamAddress := testutils.TestAddress("myteam")
	std.TestSetOrigCaller(myTeamAddress)

	alice := testutils.TestAddress("alice")

	// alice become a team
	myteam := teams.Register()

	// Setup team user address
	println("team address is team address:", myteam.IsTeamAddress(myTeamAddress))
	println("team address is not member:", !myteam.Has(myTeamAddress))
	println("team address can add package:", myteam.Can(myTeamAddress, teams.ActionAddPackage{}))
	println("alice is not member:", !myteam.Has(alice))

	println(" -> adding alice as member")
	myteam.Register(alice)

	println("alice is now a member:", myteam.Has(alice))
	println("alice cannot add a package:", !myteam.Can(alice, teams.ActionAddPackage{}))

	println(" -> setting alice AddPackage permission to true ")
	myteam.Set(alice, teams.Permissions{CanAddPackage: true})

	println("alice can now add a package:", myteam.Can(alice, teams.ActionAddPackage{}))

	println(" -> setting alice AddPackage permission to false ")
	myteam.Set(alice, teams.Permissions{CanAddPackage: false})

	println("alice cannot now add a package anymore:", !myteam.Can(alice, teams.ActionAddPackage{}))

}

// Output:
// team address is team address: true
// team address is not member: true
// team address can add package: true
// alice is not member: true
//  -> adding alice as member
// alice is now a member: true
// alice cannot add a package: true
//  -> setting alice AddPackage permission to true
// alice can now add a package: true
//  -> setting alice AddPackage permission to false
// alice cannot now add a package anymore: true
