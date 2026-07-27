// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_16_b_filetest

package z_16_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner  = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user   = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	nonMbr = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", "")
}

func main(cur realm) {
	// A non-council member cannot change listing
	testing.SetRealm(testing.NewUserRealm(nonMbr))
	commondao.SetListed(cross(cur), daoID, true)
}

// Error:
// caller is not a council member
