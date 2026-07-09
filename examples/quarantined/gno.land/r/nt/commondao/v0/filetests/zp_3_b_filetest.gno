// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_b_filetest

package zp_3_b_filetest

import (
	"testing"

	pcommondao "gno.land/p/nt/commondao/v0"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	dao    *pcommondao.CommonDAO
	subDAO *pcommondao.CommonDAO
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	dao = commondao.New(cross(cur), "Foo")
	dao.Members().Add(user)

	options := commondao.GetOptions(cross(cur), dao.ID())
	options.SetAllowDissolutionProposals(true)

	// Create a SubDAO to be dissolveed
	subDAO = commondao.NewSubDAO(cross(cur), "Bar", dao.ID())
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateDissolutionProposal(cross(cur), subDAO.ID())

	p := dao.ActiveProposals().Get(pID)
	if p == nil {
		panic("expected proposal to be created in the parent DAO")
	}

	// Make sure proposal doesn't exists in the SubDAO
	println(subDAO.ActiveProposals().Get(pID) == nil)

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Title() == ("Dissolve DAO: " + dao.Name()))
	println(p.VotingDeadline())
	println(p.Body())
}

// Output:
// true
// active
// true
// false
// 2009-02-20 23:31:30 +0000 UTC
// **DAO:**
// [Bar](/r/nt/commondao/v0:3)
