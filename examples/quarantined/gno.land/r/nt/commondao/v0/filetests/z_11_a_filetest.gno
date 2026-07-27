// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_11_a_filetest

package z_11_a_filetest

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
	pID   uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", commondao.AllowTreasuryProposals(true))

	testing.IssueCoins(commondao.GetView(daoID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 1000)))

	pID = commondao.CreateTreasurySpendProposal(cross(cur), daoID, owner, "ugnot", 400)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	commondao.Execute(cross(cur), daoID, pID)

	view := commondao.GetView(daoID)
	p, _ := view.GetProposal(pID)
	println("status:", string(p.Status()))

	b := banker.NewReadonlyBanker()
	println("treasury:", b.GetCoins(view.Address()).String())
	println("recipient:", b.GetCoins(owner).String())
}

// Output:
// status: executed
// treasury: 600ugnot
// recipient: 400ugnot
