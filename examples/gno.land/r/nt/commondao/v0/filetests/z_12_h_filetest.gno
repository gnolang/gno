// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_h_filetest

package z_12_h_filetest

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
	rootID  uint64
	subID   uint64
	spendID uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subID = 3

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 500)))

	// Pass a spend on the sub-DAO but do not execute it yet
	spendID = commondao.CreateTreasurySpendProposal(cross(cur), subID, owner, "ugnot", 500)
	commondao.Vote(cross(cur), subID, spendID, pdao.ChoiceYes, "")

	// The parent claws back the full balance before the spend executes
	cID := commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, subID)
	commondao.Vote(cross(cur), rootID, cID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, cID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The standing passed spend fails cleanly on the drained treasury
	commondao.Execute(cross(cur), subID, spendID)

	p, _ := commondao.GetView(subID).GetProposal(spendID)
	println("status:", string(p.Status()))
	println("reason:", p.StatusReason())
	b := banker.NewReadonlyBanker()
	println("sub treasury:", b.GetCoins(commondao.GetView(subID).Address()).String())
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())
	println("recipient:", b.GetCoins(owner).String())
}

// Output:
// status: failed
// reason: insufficient treasury balance
// sub treasury:
// root treasury: 500ugnot
// recipient:
