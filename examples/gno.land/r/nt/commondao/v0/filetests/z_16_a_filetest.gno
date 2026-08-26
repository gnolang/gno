// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_16_a_filetest

package z_16_a_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	// Listing defaults off
	println("default:", commondao.IsListed(daoID))

	// A council member lists the DAO
	testing.SetRealm(testing.NewUserRealm(user))
	commondao.SetListed(cross(cur), daoID, true)
	println("after list:", commondao.IsListed(daoID))

	// ...and can unlist it
	commondao.SetListed(cross(cur), daoID, false)
	println("after unlist:", commondao.IsListed(daoID))
}

// Output:
// default: false
// after list: true
// after unlist: false
