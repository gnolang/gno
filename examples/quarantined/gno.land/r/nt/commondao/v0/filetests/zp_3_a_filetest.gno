// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_a_filetest

package zp_3_a_filetest

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
	daoID = commondao.New(cross(cur), "Foo", string(user))

	commondao.UpdateOptions(cross(cur), daoID, commondao.AllowDissolutionProposals(true))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateDissolutionProposal(cross(cur), daoID)

	p, found := commondao.GetView(daoID).GetProposal(pID)
	if !found {
		panic("expected proposal to be created")
	}

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Title() == ("Dissolve DAO: " + commondao.GetView(daoID).Name()))
	println(p.VotingDeadline())
	println(p.Body())
}

// Output:
// active
// true
// true
// 2009-02-20 23:31:30 +0000 UTC
// **DAO:**
// [Foo](/r/nt/commondao/v0:2)
