// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_a_filetest

package z_12_a_filetest

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
	subID  uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID = commondao.New(cross(cur), "Root", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subID = 3

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 750)))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The ancestor claws back the sub-DAO treasury at simple majority
	pID := commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, subID)
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)

	b := banker.NewReadonlyBanker()
	println("sub treasury:", b.GetCoins(commondao.GetView(subID).Address()).String())
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())

	// Clawing back an already-empty treasury executes as a no-op
	pID = commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, subID)
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)
	p, _ := commondao.GetView(rootID).GetProposal(pID)
	println("empty clawback:", string(p.Status()))
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())
}

// Output:
// sub treasury:
// root treasury: 750ugnot
// empty clawback: executed
// root treasury: 750ugnot
