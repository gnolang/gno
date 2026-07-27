// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_14_a_filetest

package z_14_a_filetest

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
	rootID = commondao.New(cross(cur), "Root", "",
		commondao.AllowChildren(true),
		commondao.AllowTreasuryProposals(true),
	)
	subID = commondao.NewSubDAO(cross(cur), "Sub", rootID, "", commondao.AllowTreasuryProposals(true))

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 100)))

	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, true)
	commondao.Vote(cross(cur), rootID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, fID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	println("frozen:", commondao.GetView(subID).IsTreasuryFrozen())

	// The ancestor unfreezes the treasury, re-enabling spends
	uID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, false)
	commondao.Vote(cross(cur), rootID, uID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, uID)
	println("frozen:", commondao.GetView(subID).IsTreasuryFrozen())

	sID := commondao.CreateTreasurySpendProposal(cross(cur), subID, owner, "ugnot", 100)
	commondao.Vote(cross(cur), subID, sID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), subID, sID)
	println("recipient:", banker.NewReadonlyBanker().GetCoins(owner).String())
}

// Output:
// frozen: true
// frozen: false
// recipient: 100ugnot
