// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_11_e_filetest

package z_11_e_filetest

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
	rootID = commondao.New(cross(cur), "Root", "",
		commondao.AllowChildren(true),
		commondao.AllowTreasuryProposals(true),
	)
	subID = commondao.NewSubDAO(cross(cur), "Sub", rootID, "", commondao.AllowTreasuryProposals(true))

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 500)))

	// Pass a spend on the sub-DAO but do not execute it yet
	spendID = commondao.CreateTreasurySpendProposal(cross(cur), subID, owner, "ugnot", 500)
	commondao.Vote(cross(cur), subID, spendID, pdao.ChoiceYes, "")

	// The ancestor freezes the sub-DAO treasury before the spend executes
	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, true)
	commondao.Vote(cross(cur), rootID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, fID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The standing passed spend fails cleanly once the treasury is frozen
	commondao.Execute(cross(cur), subID, spendID)

	view := commondao.GetView(subID)
	p, _ := view.GetProposal(spendID)
	println("status:", string(p.Status()))
	println("reason:", p.StatusReason())
	println("frozen:", view.IsTreasuryFrozen())
	println("treasury:", banker.NewReadonlyBanker().GetCoins(view.Address()).String())
}

// Output:
// status: failed
// reason: DAO treasury is frozen
// frozen: true
// treasury: 500ugnot
