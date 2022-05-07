// PKGPATH: gno.land/r/gnotter_tests
package gnotter_tests

import (
	"gno.land/r/gnotter"
)

var gProfileID gnotter.ProfileID
var gAnotherProfileID gnotter.ProfileID

func init() {
	gProfileID = gnotter.CreateProfile("Milerius")
	gAnotherProfileID = gnotter.CreateProfile("Slyris")
}

func main() {
	println("-----------")
	println(gnotter.Render("Milerius"))
	println("-----------")
	println(gnotter.Render("Slyris"))
	println("-----------")
	gnotter.Follow(gProfileID, gAnotherProfileID)
	println(gnotter.Render("Milerius"))
	println("-----------")
	println(gnotter.Render("Slyris"))
	println("-----------")
	gnotter.UnFollow(gProfileID, gAnotherProfileID)
	println(gnotter.Render("Milerius"))
	println("-----------")
	println(gnotter.Render("Slyris"))
	println("-----------")
}

// Output:
// -----------
// Name: Milerius
// URL: /r/profile:Milerius
// Following: 0
// Follower: 0
// Nb threads: 0
// -----------
// Name: Slyris
// URL: /r/profile:Slyris
// Following: 0
// Follower: 0
// Nb threads: 0
// -----------
// Name: Milerius
// URL: /r/profile:Milerius
// Following: 1
// Follower: 0
// Nb threads: 0
// -----------
// Name: Slyris
// URL: /r/profile:Slyris
// Following: 0
// Follower: 1
// Nb threads: 0
// -----------
// Name: Milerius
// URL: /r/profile:Milerius
// Following: 0
// Follower: 0
// Nb threads: 0
// -----------
// Name: Slyris
// URL: /r/profile:Slyris
// Following: 0
// Follower: 0
// Nb threads: 0
// -----------
