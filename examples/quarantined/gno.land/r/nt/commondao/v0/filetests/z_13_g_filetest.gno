// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_13_g_filetest

package z_13_g_filetest

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
	rootID = commondao.New(cross(cur), "Root", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subID = 3

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 700)))

	// Pass a spend on the sub-DAO but do not execute it
	spendID = commondao.CreateTreasurySpendProposal(cross(cur), subID, owner, "ugnot", 700)
	commondao.Vote(cross(cur), subID, spendID, pdao.ChoiceYes, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Dissolution dismisses the passed-but-unexecuted spend and sweeps the
	// full balance to the parent; the recipient never receives anything
	dID := commondao.CreateDissolutionProposal(cross(cur), subID, "")
	commondao.Vote(cross(cur), rootID, dID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, dID)

	p, _ := commondao.GetView(subID).GetProposal(spendID)
	println("spend:", string(p.Status()))
	println("reason:", p.StatusReason())
	b := banker.NewReadonlyBanker()
	println("sub treasury:", b.GetCoins(commondao.GetView(subID).Address()).String())
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())
	println("recipient:", b.GetCoins(owner).String())
}

// Output:
// spend: dismissed
// reason: DAO dissolved
// sub treasury:
// root treasury: 700ugnot
// recipient:
