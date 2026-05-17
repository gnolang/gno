// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_1_d_filetest

package zp_1_d_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var dao *pcommondao.CommonDAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross2(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross2(cur), "Foo")
	dao.Members().Add(user)
	dao.Members().Add(owner)

	options := commondao.GetOptions(cross2(cur), dao.ID())
	options.SetAllowTextProposals(true)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.CreateTextProposal(cross2(cur), dao.ID(), "", "", 31)
}

// Error:
// maximum proposal voting period is 30 days
