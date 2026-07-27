// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_13_e_filetest

package z_13_e_filetest

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

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "Foo", "", "")

	testing.IssueCoins(commondao.GetView(daoID).Address(), chain.NewCoins(chain.NewCoin("ugnot", 650)))
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	pID := commondao.CreateDissolutionProposal(cross(cur), daoID, owner)
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)

	view := commondao.GetView(daoID)
	println("deleted:", view.IsDeleted())
	b := banker.NewReadonlyBanker()
	println("treasury:", b.GetCoins(view.Address()).String())
	println("destination:", b.GetCoins(owner).String())
}

// Output:
// deleted: true
// treasury:
// destination: 650ugnot
