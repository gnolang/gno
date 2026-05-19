// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_e_filetest

package zp_3_e_filetest

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
	dao := commondao.New(cross(cur), "Foo")

	options := commondao.GetOptions(cross(cur), dao.ID())
	options.SetAllowDissolutionProposals(true)

	daoID = dao.ID()
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateDissolutionProposal(cross(cur), daoID)
}

// Error:
// caller is not a DAO member
