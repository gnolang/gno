// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_2_f_filetest

package zp_2_f_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	name  = "Foo"
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao := commondao.New(cross(cur), "Foo")
	dao.Members().Add(user)

	commondao.NewSubDAO(cross(cur), name, dao.ID())

	options := commondao.GetOptions(cross(cur), dao.ID())
	options.SetAllowChildren(true)
	options.SetAllowSubDAOPorposals(true)

	daoID = dao.ID()
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateSubDAOProposal(cross(cur), daoID, name, "")
}

// Error:
// a SubDAO with the same name already exists
