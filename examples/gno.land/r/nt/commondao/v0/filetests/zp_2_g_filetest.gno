// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_2_g_filetest

package zp_2_g_filetest

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
	commondao.Invite(cross2(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross2(cur), "Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(cross2(cur), dao.ID())
	options.SetAllowChildren(true)
	options.SetAllowSubDAOPorposals(true)

	daoID = dao.ID()
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateSubDAOProposal(cross2(cur), daoID, "", "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
}

// Error:
// DAO name is empty
