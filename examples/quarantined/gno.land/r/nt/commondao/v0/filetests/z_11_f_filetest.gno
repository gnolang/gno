// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_11_f_filetest

package z_11_f_filetest

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
	daoID uint64
	p1    uint64
	p2    uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", "")

	testing.IssueCoins(commondao.GetView(daoID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	// Two passed spends over-commit the 1000ugnot balance
	p1 = commondao.CreateTreasurySpendProposal(cross(cur), daoID, owner, "ugnot", 800)
	commondao.Vote(cross(cur), daoID, p1, pdao.ChoiceYes, "")
	p2 = commondao.CreateTreasurySpendProposal(cross(cur), daoID, owner, "ugnot", 800)
	commondao.Vote(cross(cur), daoID, p2, pdao.ChoiceYes, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Execute(cross(cur), daoID, p1)
	commondao.Execute(cross(cur), daoID, p2)

	view := commondao.GetView(daoID)
	v1, _ := view.GetProposal(p1)
	v2, _ := view.GetProposal(p2)
	println("first:", string(v1.Status()))
	println("second:", string(v2.Status()), "-", v2.StatusReason())
	println("treasury:", banker.NewReadonlyBanker().GetCoins(view.Address()).String())
}

// Output:
// first: executed
// second: failed - insufficient treasury balance
// treasury: 200ugnot
