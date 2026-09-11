// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_14_e_filetest

package z_14_e_filetest

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

	testing.IssueCoins(commondao.GetView(leafID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 100)))

	// The middle DAO freezes the leaf, then the root dissolves the middle
	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), midID, leafID, true)
	commondao.Vote(cross(cur), midID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), midID, fID)

	dID := commondao.CreateDissolutionProposal(cross(cur), midID, "")
	commondao.Vote(cross(cur), rootID, dID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, dID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Freeze state lives on the target: any live proper ancestor can
	// unfreeze it after the freezing ancestor is gone
	uID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, leafID, false)
	commondao.Vote(cross(cur), rootID, uID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, uID)
	println("frozen:", commondao.GetView(leafID).IsTreasuryFrozen())

	sID := commondao.CreateTreasurySpendProposal(cross(cur), leafID, owner, "ugnot", 100)
	commondao.Vote(cross(cur), leafID, sID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), leafID, sID)
	println("recipient:", banker.NewReadonlyBanker().GetCoins(owner).String())
}

// Output:
// frozen: false
// recipient: 100ugnot
