// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_e_filetest

package z_12_e_filetest

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
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Mid", "Purpose", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	midID = 3
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), midID, "Leaf", "Purpose", "", string(user))
		commondao.Vote(cross(cur), midID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), midID, pID)
	}
	leafID = 4

	testing.IssueCoins(commondao.GetView(leafID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 420)))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The grandparent proposes; swept funds must land on the target's
	// parent (one step up the tree), never on the proposing DAO
	pID := commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, leafID)
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)

	b := banker.NewReadonlyBanker()
	println("leaf treasury:", b.GetCoins(commondao.GetView(leafID).Address()).String())
	println("mid treasury:", b.GetCoins(commondao.GetView(midID).Address()).String())
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())
}

// Output:
// leaf treasury:
// mid treasury: 420ugnot
// root treasury:
