// PKGPATH: gno.land/r/nt/commondao/v0/filetests/zp_3_b_filetest

package zp_3_b_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoID uint64
	subID uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/demo/test"))
	daoID = commondao.New(cross(cur), "Foo", string(user))

	commondao.UpdateOptions(cross(cur), daoID, commondao.AllowDissolutionProposals(true))

	// Create a SubDAO to be dissolved
	subID = commondao.NewSubDAO(cross(cur), "Bar", daoID, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateDissolutionProposal(cross(cur), subID)

	p, found := commondao.GetView(daoID).GetProposal(pID)
	if !found {
		panic("expected proposal to be created in the parent DAO")
	}

	// Make sure proposal doesn't exists in the SubDAO
	_, foundInSub := commondao.GetView(subID).GetProposal(pID)
	println(!foundInSub)

	println(string(p.Status()))
	println(p.Creator() == user)
	println(p.Title() == ("Dissolve DAO: " + commondao.GetView(subID).Name()))
	println(p.VotingDeadline())
	println(p.Body())
}

// Output:
// true
// active
// true
// true
// 2009-02-20 23:31:30 +0000 UTC
// **DAO:**
// [Bar](/r/nt/commondao/v0:3)
