// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_13_d_filetest

package z_13_d_filetest

import (
	"chain"
	"chain/banker"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"
	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	rootID uint64
	midID  uint64
	leafID uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID = commondao.New(cross(cur), "Root", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Mid", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	midID = 3
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), midID, "Leaf", "", string(user))
		commondao.Vote(cross(cur), midID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), midID, pID)
	}
	leafID = 4

	testing.IssueCoins(commondao.GetView(leafID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 150)))

	// Dissolve the middle DAO first, orphaning the leaf
	pID := commondao.CreateDissolutionProposal(cross(cur), midID, "")
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The leaf's dissolution proposal is hosted in the nearest live
	// ancestor (the root), because its parent is already dissolved
	pID := commondao.CreateDissolutionProposal(cross(cur), leafID, "")
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)

	println("mid deleted:", commondao.GetView(midID).IsDeleted())
	println("leaf deleted:", commondao.GetView(leafID).IsDeleted())

	// The orphan's sweep goes to its (deleted) parent's treasury, where
	// the funds stay clawback-rescuable by the live root
	b := banker.NewReadonlyBanker()
	println("leaf treasury:", b.GetCoins(commondao.GetView(leafID).Address()).String())
	println("mid treasury:", b.GetCoins(commondao.GetView(midID).Address()).String())
}

// Output:
// mid deleted: true
// leaf deleted: true
// leaf treasury:
// mid treasury: 150ugnot
