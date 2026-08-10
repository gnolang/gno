// PKGPATH: gno.land/r/demo/test
package test

import (
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/p/nt/testutils/v0"

	"gno.land/r/nt/commondao/v0"
)

const owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx

var (
	daoID uint64
	pID   uint64
	user1 = testutils.TestAddress("user1")
)

func init(cur realm) {
	// Invite a user to be able to start creating DAOs
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user1)

	// Create a new DAO owned by user1, with user1 as the only council member
	testing.SetRealm(testing.NewUserRealm(user1))
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	// Create a new proposal
	pID = commondao.CreateTextProposal(cross(cur), daoID, "Title", "Body", 0)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user1))

	// A single YES is a supermajority of the one member electorate,
	// so the proposal passes immediately and can be executed early
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)

	p, found := commondao.GetView(daoID).GetProposal(pID)
	if !found {
		panic("expected proposal to be finished")
	}

	println(string(p.Status()))
	println(commondao.GetView(daoID).FinishedProposalsSize())
}

// Output:
// executed
// 1
