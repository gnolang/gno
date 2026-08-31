// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_i_filetest

package z_12_i_filetest

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

	testing.IssueCoins(commondao.GetView(leafID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 900)))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Root and mid both pass clawbacks against the same leaf
	rID := commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, leafID)
	commondao.Vote(cross(cur), rootID, rID, pdao.ChoiceYes, "")
	mID := commondao.CreateTreasuryClawbackProposal(cross(cur), midID, leafID)
	commondao.Vote(cross(cur), midID, mID, pdao.ChoiceYes, "")

	// Funds move exactly once, to the leaf's parent; the second clawback
	// executes as an empty no-op and does not hoist mid's new balance
	commondao.Execute(cross(cur), rootID, rID)
	commondao.Execute(cross(cur), midID, mID)

	rp, _ := commondao.GetView(rootID).GetProposal(rID)
	mp, _ := commondao.GetView(midID).GetProposal(mID)
	println("root clawback:", string(rp.Status()))
	println("mid clawback:", string(mp.Status()))
	b := banker.NewReadonlyBanker()
	println("leaf treasury:", b.GetCoins(commondao.GetView(leafID).Address()).String())
	println("mid treasury:", b.GetCoins(commondao.GetView(midID).Address()).String())
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())
}

// Output:
// root clawback: executed
// mid clawback: executed
// leaf treasury:
// mid treasury: 900ugnot
// root treasury:
