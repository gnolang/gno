// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_13_f_filetest

package z_13_f_filetest

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
	rootID = commondao.New(cross(cur), "Root", "Purpose", "", "")
	{
		pID := commondao.CreateSubDAOProposal(cross(cur), rootID, "Sub", "Purpose", "", string(user))
		commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
		commondao.Execute(cross(cur), rootID, pID)
	}
	subID = 3

	testing.IssueCoins(commondao.GetView(subID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 800)))

	fID := commondao.CreateTreasuryFreezeProposal(cross(cur), rootID, subID, true)
	commondao.Vote(cross(cur), rootID, fID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, fID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// A frozen treasury does not block dissolution: the sweep sends the
	// funds where a clawback would put them
	pID := commondao.CreateDissolutionProposal(cross(cur), subID, "")
	commondao.Vote(cross(cur), rootID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), rootID, pID)

	view := commondao.GetView(subID)
	println("deleted:", view.IsDeleted())
	println("frozen:", view.IsTreasuryFrozen())
	b := banker.NewReadonlyBanker()
	println("sub treasury:", b.GetCoins(view.Address()).String())
	println("root treasury:", b.GetCoins(commondao.GetView(rootID).Address()).String())
}

// Output:
// deleted: true
// frozen: true
// sub treasury:
// root treasury: 800ugnot
