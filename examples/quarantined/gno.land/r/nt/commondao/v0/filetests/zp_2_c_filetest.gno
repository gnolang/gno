// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_2_c_filetest

package zp_2_c_filetest

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
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross(cur), "Foo", "", "")

	// Disallow SubDAO creation
	commondao.UpdateOptions(cross(cur), daoID, false, true, false, true, 0)

}

func main(cur realm) {
	commondao.CreateSubDAOProposal(cross(cur), daoID, "", "", "")
}

// Error:
// SubDAO support is not enabled
