// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_12_j_filetest

package z_12_j_filetest

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

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(
		chain.NewCoin("ugnot", 100),
		chain.NewCoin("zzfoo", 50),
	))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// Clawback sweeps the full multi-denomination balance
	pID := commondao.CreateTreasuryClawbackProposal(cross(cur), rootID, subID)
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)

	b := banker.NewReadonlyBanker()
	println("sub treasury:", b.GetCoins(commondao.GetView(subID).Address()).String())
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())
}

// Output:
// sub treasury:
// root treasury: 100ugnot,50zzfoo
