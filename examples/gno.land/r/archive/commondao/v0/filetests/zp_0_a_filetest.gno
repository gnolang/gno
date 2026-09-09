// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_0_a_filetest

package zp_0_a_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner         = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user          = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
	newMembers    = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq\ng147ah9520z0r6jh9mjr6c75rv6l8aypzvcd3f7d"
	removeMembers = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
)

var dao *pcommondao.CommonDAO

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross(cur), "Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(cross(cur), dao.ID())
	options.SetAllowMembersUpdate(true)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateMembersUpdateProposal(cross(cur), dao.ID(), newMembers, removeMembers)

	p := dao.ActiveProposals().Get(pID)
	if p == nil {
		panic("expected proposal to be created")
	}

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Definition().Title() == "Members Update")
	println("")
	println(p.Definition().Body())
}

// Output:
// active
// true
// true
//
// **Members to Add:**
// - g147ah9520z0r6jh9mjr6c75rv6l8aypzvcd3f7d
// - g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq
//
//
// **Members to Remove:**
// - g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
//
