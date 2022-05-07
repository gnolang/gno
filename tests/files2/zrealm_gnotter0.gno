// PKGPATH: gno.land/r/gnotter_tests
package gnotter_tests

import (
	"gno.land/r/gnotter"
)

var gProfileID gnotter.ProfileID

func init() {
	gProfileID = gnotter.CreateProfile("Milerius")
}

func main() {
	println("-----------")
	println(gnotter.Render("Milerius"))
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
