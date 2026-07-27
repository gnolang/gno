// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_16_d_filetest

package z_16_d_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	listedID := commondao.New(cross(cur), "Shown", "Purpose", "", "")
	commondao.New(cross(cur), "Hidden", "Purpose", "", "")

	commondao.SetListed(cross(cur), listedID, true)
}

func main() {
	// The home index shows only listed DAOs: "Shown" appears, "Hidden"
	// (default unlisted) does not.
	println(commondao.Render(""))
}

// Output:
// # Common DAO
// ---
// This realm can be used to create CommonDAO instances based on [commondao](/p/nt/commondao/v0/) package.
//
// Here is a list of some of the DAOs that were created:
//
// - [Shown](/r/nt/commondao/v0:2)
// - [Common DAO](/r/nt/commondao/v0:1)
